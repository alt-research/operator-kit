// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package awstools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/alt-research/operator-kit/must"
	dockertypes "github.com/docker/docker/api/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type dockerRegAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func EcrDockerClientAuth(ctx context.Context, svc ...*ecr.Client) dockertypes.RequestPrivilegeFunc {
	var c *ecr.Client
	if len(svc) > 0 {
		c = svc[0]
	}
	return func() (string, error) {
		u, p, err := GetECRAuthToken(ctx, c)
		if err != nil {
			return "", err
		}
		j := must.Two(json.Marshal(dockerRegAuth{Username: u, Password: p}))
		token := base64.StdEncoding.EncodeToString(j)
		return token, nil
	}
}

func GetECRAuthToken(ctx context.Context, svc *ecr.Client) (string, string, error) {
	log := log.FromContext(ctx)
	if svc == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.V(1).Error(err, "failed to load default aws config")
			return "", "", err
		}

		svc = ecr.NewFromConfig(cfg)
	}

	token, err := svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.V(1).Error(err, "failed to fetch ECR authorization token")
		return "", "", err
	}
	basic, err := base64.StdEncoding.DecodeString(*token.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		log.V(1).Error(err, "failed to decode ECR authorization token")
	}
	split := strings.Split(string(basic), ":")
	return split[0], split[1], nil
}
