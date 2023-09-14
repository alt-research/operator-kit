// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package k8s

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/alt-research/operator-kit/must"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	clientset *kubernetes.Clientset
	home      = homedir.HomeDir()
	cfgPath   = must.Default(os.Getenv("KUBECONFIG"), filepath.Join(home, ".kube", "config"))
)

func GetClient() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}
	config, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func Logs(ctx context.Context, namespace, pod, container string) (string, error) {
	clientset, err := GetClient()
	if err != nil {
		return "", err
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{Container: container})
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer readCloser.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, readCloser)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetSelfServiceAccount(ctx context.Context, namespace string) (string, error) {
	if stat, _ := os.Stat(cfgPath); stat != nil && stat.IsDir() {
		return os.Getenv("SERVICE_ACCOUNT"), nil
	}
	clientset, err := GetClient()
	if err != nil {
		return "", err
	}
	namespace = must.Default(namespace, "default")
	name := os.Getenv("POD_NAME")
	if name == "" {
		return os.Getenv("SERVICE_ACCOUNT"), nil
	}
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return pod.Spec.ServiceAccountName, nil
}
