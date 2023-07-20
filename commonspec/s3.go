package commonspec

import (
	"fmt"
	"os"
	"time"

	"github.com/alt-research/operator-kit/must"
	"github.com/alt-research/operator-kit/s3util"
	"github.com/dustin/go-humanize"
)

type S3ObjectRef struct {
	//+optional
	Endpoint string `json:"endpoint,omitempty"`
	//+required
	Region string `json:"region"`
	//+required
	Bucket string `json:"bucket"`
	// The exact object key of the newly created object.
	// if provided, will use this key to upload/get the the object
	// if empty, normally the key will be generated by the operator as KeyPrefix/Filename
	Key string `json:"key,omitempty"`
	// Key Prefix of the newly created object.
	KeyPrefix string `json:"KeyPrefix,omitempty"`
	// Filename is the postfix of the object key.
	Filename string `json:"filename,omitempty"`
	// The URL where the object was uploaded to.
	Location string `json:"location,omitempty"`

	Size uint64 `json:"size,omitempty"`

	// The ID for a multipart upload to S3. In the case of an error the error
	// can be cast to the MultiUploadFailure interface to extract the upload ID.
	// Will be empty string if multipart upload was not used, and the object
	// was uploaded as a single PutObject call.
	UploadID string `json:"uploadId,omitempty"`

	// Entity tag for the uploaded object.
	ETag *string `json:"eTag,omitempty"`

	// If the object expiration is configured, this will contain the expiration date
	// (expiry-date) and rule ID (rule-id). The value of rule-id is URL encoded.
	Expiration *string `json:"expiration,omitempty"`

	// The version of the object that was uploaded. Will only be populated if
	// the S3 Bucket is versioned. If the bucket is not versioned this field
	// will not be set.
	VersionID *string `json:"versionId,omitempty"`

	//+kubebuilder:validation:Format="date-time"
	LastModified string `json:"lastModified,omitempty"`

	//+kubebuilder:validation:Enum=STANDARD;REDUCED_REDUNDANCY;STANDARD_IA;ONEZONE_IA;INTELLIGENT_TIERING;GLACIER;DEEP_ARCHIVE;OUTPOSTS
	//+optional
	StorageClass string `json:"storageClass,omitempty"`
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl
	//+kubebuilder:validation:Enum=private;public-read;public-read-write;authenticated-read;aws-exec-read;bucket-owner-read;bucket-owner-full-control
	ObjectACL string `json:"objectACL,omitempty"`
}

type S3ObjectRefStatus struct {
	Url   string `json:"url,omitempty"`
	S3Url string `json:"s3Url,omitempty"`
	Size  string `json:"size,omitempty"`
}

func (o *S3ObjectRef) Url() string {
	if o.Location != "" {
		return o.Location
	}
	return fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", o.Region, o.Bucket, o.Key)
}

func (o *S3ObjectRef) S3Url() string {
	return fmt.Sprintf("s3://%s/%s", o.Bucket, o.Key)
}

func (o *S3ObjectRef) SizeReadable() string {
	return humanize.IBytes(o.Size)
}

func (o *S3ObjectRef) FromUpload(up s3util.UploadOutput) error {
	o.Location = up.Location
	// u, err := url.Parse(up.Location)
	// if err != nil {
	// 	return err
	// }
	// split := strings.Split(u.Host, ".")
	// if len(split) < 3 || !array.Contains(awstools.Regions, split[1]) {
	// 	o.Region = must.Default(os.Getenv("AWS_REGION"), os.Getenv("AWS_DEFAULT_REGION"))
	// } else {
	// 	o.Region = split[1]
	// }
	o.Endpoint = up.Endpoint
	o.Region = up.Region
	o.Bucket = up.Bucket
	o.Size = uint64(up.Size)
	// o.Bucket = strings.TrimSuffix(u.Path[1:], "/"+*up.Key)
	o.Key = *up.Key
	o.ETag = up.ETag
	o.Expiration = up.Expiration
	o.UploadID = up.UploadID
	o.VersionID = up.VersionID
	return nil
}

func (o *S3ObjectRef) FromHeadOutput(head s3util.HeadObjectOutput) {
	o.Key = head.Key
	o.Endpoint = head.Endpoint
	o.Bucket = head.Bucket
	o.Region = head.Region
	o.ETag = head.ETag
	o.Expiration = head.Expiration
	o.StorageClass = string(head.StorageClass)
	o.LastModified = head.LastModified.Format(time.RFC3339)
	o.VersionID = head.VersionId
	o.Size = uint64(head.ContentLength)
}

func (o *S3ObjectRef) ToStatus(s *S3ObjectRefStatus) {
	s.Url = o.Url()
	s.S3Url = o.S3Url()
	s.Size = o.SizeReadable()
}

func (r *S3ObjectRef) SetDefaults() {
	r.Region = must.Default(os.Getenv("AWS_REGION"), os.Getenv("AWS_DEFAULT_REGION"))
	r.ObjectACL = "private"
	if r.Key == "" {
		r.Key = r.KeyPrefix + "/" + r.Filename
	}
}
