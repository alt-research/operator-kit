package specutil

import (
	"os"

	"github.com/kataras/go-fs"
)

func IsWebhookReady() bool {
	return fs.DirectoryExists("/tmp/k8s-webhook-server/serving-certs/tls.crt") || os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}
