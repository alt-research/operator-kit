// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package s3util

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alt-research/operator-kit/awstools"
	"github.com/alt-research/operator-kit/ptr"
	"github.com/alt-research/operator-kit/syncmap"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3mgr "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BucketManager struct {
	aws.Config
	Opts        awstools.AWSCfgOpts
	Bucket      string
	Prefix      string
	concurrency int
	client      *awss3.Client
	uploader    *s3mgr.Uploader
	downloader  *s3mgr.Downloader

	sem *semaphore.Weighted
}

func NewManagerWithClient(client *awss3.Client, bucket, prefix string, concurrency int) (*BucketManager, error) {
	if concurrency == 0 {
		concurrency = 5
	}
	return &BucketManager{
		Bucket:      bucket,
		Prefix:      prefix,
		concurrency: concurrency,
		client:      client,
		uploader: s3mgr.NewUploader(client, func(u *s3mgr.Uploader) {
			u.Concurrency = 5
		}),
		downloader: s3mgr.NewDownloader(client, func(d *s3mgr.Downloader) {
			d.Concurrency = 5
		}),
		sem: semaphore.NewWeighted(int64(concurrency)),
	}, nil
}

func NewManager(bucket, prefix string, concurrency int, opts ...awstools.AWSCfgOpts) (*BucketManager, error) {
	cfg, opt, err := awstools.GetCfg(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	client := awss3.NewFromConfig(cfg, func(o *awss3.Options) { o.UsePathStyle = true })
	c, err := NewManagerWithClient(client, bucket, prefix, concurrency)
	if err != nil {
		return nil, err
	}
	c.Config = cfg
	c.Opts = opt
	return c, err
}

func (b *BucketManager) SetConcurrency(n int) {
	b.concurrency = n
	b.sem = semaphore.NewWeighted(int64(n))
}

type UploadOptions struct {
	ACL          types.ObjectCannedACL
	StorageClass types.StorageClass
	Tagging      *string
}

type UploadOutput struct {
	s3mgr.UploadOutput
	Endpoint string
	Region   string
	Bucket   string
	Size     int64
}

type UploadOutputs struct {
	syncmap.Map[string, UploadOutput]
}

func (b *BucketManager) Upload(ctx context.Context, src, keyOrPrefix string, filter func(string) bool, opts *UploadOptions) (*UploadOutputs, error) {
	uploaded := new(UploadOutputs)
	log := log.FromContext(ctx)
	wg := &sync.WaitGroup{}
	if info, err := os.Stat(src); err == nil {
		if !info.IsDir() {
			out, err := b.UploadSingle(ctx, src, keyOrPrefix, opts)
			if err != nil || out == nil {
				return nil, err
			}
			out.Size = info.Size()
			uploaded.Store(src, *out)
			return uploaded, err
		} else {
			if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() {
					return nil
				}
				if filter != nil && !filter(path) {
					return nil
				}
				// if err := b.sem.Acquire(ctx, 1); err != nil {
				// 	return errors.Wrap(err, "Failed to acquire semaphore")
				// }
				wg.Add(1)
				go func() {
					// defer b.sem.Release(1)
					defer wg.Done()
					rel := strings.TrimPrefix(strings.ReplaceAll(path, `\`, `/`), src)
					rel = strings.TrimPrefix(rel, "/")
					key := filepath.Join(keyOrPrefix, rel)
					out, err := b.UploadSingle(ctx, path, key, opts)
					if err != nil || out == nil {
						log.Error(err, "upload failed", "path", path, "key", key, "bucket", b.Bucket)
						return
					}
					log.V(1).Info("Uploaded file", "path", path, "key", out.Key, "etag", out.ETag, "bucket", b.Bucket)
					stat, err := os.Stat(path)
					if err != nil {
						log.Error(err, "Failed to stat file", "path", path)
						return
					}
					out.Size = stat.Size()
					uploaded.Store(rel, *out)
				}()
				return nil
			}); err != nil {
				wg.Wait()
				return nil, err
			}
		}
	}
	wg.Wait()
	return uploaded, nil
}

func (b *BucketManager) UploadSingle(ctx context.Context, src, keyOrPrefix string, opts *UploadOptions) (*UploadOutput, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return b.UploadReader(ctx, src, file, keyOrPrefix, opts)
}

func (b *BucketManager) UploadReader(ctx context.Context, name string, src io.Reader, keyOrPrefix string, opts *UploadOptions) (*UploadOutput, error) {
	if err := b.sem.Acquire(ctx, 1); err != nil {
		return nil, errors.Wrap(err, "Failed to acquire semaphore")
	}
	defer b.sem.Release(1)
	key := filepath.Join(b.Prefix, keyOrPrefix)
	if keyOrPrefix == "" || keyOrPrefix[len(keyOrPrefix)-1] == '/' {
		key = filepath.Join(key, filepath.Base(name))
	} else if is, err := b.IsPathDir(ctx, key); err != nil {
		return nil, err
	} else if is {
		key = filepath.Join(key, filepath.Base(name))
	}
	input := &awss3.PutObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
		Body:   src,
	}
	if opts != nil {
		input.ACL = opts.ACL
		input.StorageClass = opts.StorageClass
		input.Tagging = opts.Tagging
	}
	up, err := b.uploader.Upload(ctx, input)
	if err != nil {
		return nil, err
	}
	return &UploadOutput{UploadOutput: *up, Endpoint: b.Opts.Endpoint, Region: b.Region, Bucket: b.Bucket}, nil
}

func (b *BucketManager) IsPathDir(ctx context.Context, path string) (bool, error) {
	if path == "" || path[len(path)-1] == '/' {
		return true, nil
	}
	res, err := b.client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: &b.Bucket, Prefix: &path, MaxKeys: ptr.Of(int32(2))})
	if err != nil {
		return false, err
	}
	for _, r := range res.Contents {
		if filepath.Dir(*r.Key) == path {
			return true, nil
		}
	}
	return false, nil
}

type HeadObjectOutput struct {
	Endpoint string
	Region   string
	Bucket   string
	Key      string
	awss3.HeadObjectOutput
}

func (b *BucketManager) HeadObject(ctx context.Context, key string) (HeadObjectOutput, error) {
	h, err := b.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
	})
	if err != nil {
		return HeadObjectOutput{}, err
	}
	return HeadObjectOutput{
		Endpoint:         b.Opts.Endpoint,
		Bucket:           b.Bucket,
		Region:           b.Region,
		Key:              key,
		HeadObjectOutput: *h,
	}, nil
}

func (b *BucketManager) DeleteObject(ctx context.Context, key string) (*awss3.DeleteObjectOutput, error) {
	return b.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
	})
}

func (b *BucketManager) Download(ctx context.Context, key string, dst string, overwrite bool) (files []string, err error) {
	if ok, err := b.IsPathDir(ctx, key); err != nil {
		return []string{}, err
	} else if !ok {
		file, err := b.DownloadSingle(ctx, key, dst, overwrite)
		return []string{file}, err
	}
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return []string{}, errors.Errorf("dest %s already exists", dst)
		}
	}
	err = os.MkdirAll(dst, 0o755)
	if err != nil {
		return []string{}, err
	}
	var keys []string
	var continuation *string
	for {
		objects, err := b.client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
			Bucket:            &b.Bucket,
			Prefix:            &key,
			ContinuationToken: continuation,
		})
		if err != nil {
			return []string{}, err
		}
		for _, obj := range objects.Contents {
			keys = append(keys, *obj.Key)
		}
		if !(objects.IsTruncated != nil && *objects.IsTruncated) {
			break
		}
		continuation = objects.NextContinuationToken
	}
	errG, ctx := errgroup.WithContext(ctx)
	filenameCh := make(chan string, len(keys))
	for _, k := range keys {
		k := k
		errG.Go(func() error {
			if err := b.sem.Acquire(ctx, 1); err != nil {
				return errors.Wrap(err, "Failed to acquire semaphore")
			}
			defer b.sem.Release(1)
			filename, err := b.DownloadSingle(ctx, k, dst, overwrite)
			if err != nil {
				log.Log.Error(err, "Failed to download file", "key", k)
				return err
			}
			filenameCh <- filename
			return nil
		})
	}
	go func() {
		for filename := range filenameCh {
			files = append(files, filename)
		}
	}()
	err = errG.Wait()
	close(filenameCh)
	return
}

func (b *BucketManager) DownloadSingle(ctx context.Context, key string, dst string, overwrite bool) (string, error) {
	// if dest is dir, create file in this dir with the filename as the filename of key
	if info, err := os.Stat(dst); err == nil && info.IsDir() {
		dst = filepath.Join(dst, filepath.Base(key))
	}
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return "", errors.Errorf("dest %s already exists", dst)
		}
	}
	file, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = b.DownloadWriter(ctx, key, file)
	return dst, err
}

func (b *BucketManager) DownloadWriter(ctx context.Context, key string, dst io.WriterAt) (int64, error) {
	return b.downloader.Download(ctx, dst, &awss3.GetObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
	})
}

func (b *BucketManager) Delete(ctx context.Context, key string) (deletes *awss3.DeleteObjectsOutput, err error) {
	var keys []types.ObjectIdentifier
	var continuation *string
	for {
		objects, err := b.client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
			Bucket:            &b.Bucket,
			Prefix:            &key,
			ContinuationToken: continuation,
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range objects.Contents {
			keys = append(keys, types.ObjectIdentifier{Key: obj.Key})
		}
		if !(objects.IsTruncated != nil && *objects.IsTruncated) {
			break
		}
		continuation = objects.NextContinuationToken
	}
	if len(keys) == 0 {
		return
	}
	length := len(keys)
	out := &awss3.DeleteObjectsOutput{}
	for i := 0; i < length; i += 1000 {
		end := i + 1000
		if end > length {
			end = length
		}
		res, err := b.client.DeleteObjects(ctx, &awss3.DeleteObjectsInput{
			Bucket: &b.Bucket,
			Delete: &types.Delete{
				Objects: keys[i:end],
			},
		})
		if err != nil {
			return nil, err
		}
		out.Deleted = append(out.Deleted, res.Deleted...)
		out.Errors = append(out.Errors, res.Errors...)
		out.RequestCharged = res.RequestCharged
		out.ResultMetadata = res.ResultMetadata
	}
	return out, nil
}

func (b *BucketManager) DeleteSingle(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
	})
	return err
}

func (b *BucketManager) ObjectS3URL(key string) string {
	return fmt.Sprintf("s3://%s/%s", b.Bucket, key)
}
