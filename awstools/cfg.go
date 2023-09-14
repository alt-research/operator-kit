// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package awstools

import (
	"context"
	"os"

	"github.com/alt-research/operator-kit/must"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	endpointEnvKey         = "AWS_ENDPOINT"
	regionEnvVar           = "AWS_REGION"
	awsDefaultRegionEnvVar = "AWS_DEFAULT_REGION"
)

var (
	EndpointEnvKeys = []string{endpointEnvKey}
	RegionEnvKeys   = []string{
		regionEnvVar,
		awsDefaultRegionEnvVar,
	}
)

func SetStringFromEnvVal(keys []string, _default ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); len(v) > 0 {
			return v
		}
	}
	if len(_default) > 0 {
		return _default[0]
	}
	return ""
}

type AWSCfgOpts struct {
	aws.Credentials
	Endpoint string
	Region   string
}

func GetCfg(ctx context.Context, opts ...AWSCfgOpts) (aws.Config, AWSCfgOpts, error) {
	var opt AWSCfgOpts
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt.Endpoint = must.Default(opt.Endpoint, SetStringFromEnvVal(EndpointEnvKeys))
	opt.Region = must.Default(opt.Region, SetStringFromEnvVal(RegionEnvKeys))

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if opt.Endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           opt.Endpoint,
				SigningRegion: region,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	optFns := []func(*awsCfg.LoadOptions) error{
		awsCfg.WithRegion(opt.Region),
		awsCfg.WithEndpointResolverWithOptions(customResolver),
	}
	if opt.Credentials.SecretAccessKey != "" && opt.Credentials.AccessKeyID != "" {
		optFns = append(optFns, awsCfg.WithCredentialsProvider(credentials.StaticCredentialsProvider{Value: opt.Credentials}))
	}
	cfg, err := awsCfg.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return aws.Config{}, opt, err
	}
	return cfg, opt, nil
}
