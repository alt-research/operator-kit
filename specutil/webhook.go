// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	"os"

	"github.com/kataras/go-fs"
)

func IsWebhookReady() bool {
	return fs.DirectoryExists("/tmp/k8s-webhook-server/serving-certs/tls.crt") || os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}
