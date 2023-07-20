package specutil

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func (v *ValueOrConfigMap) Get(ctx context.Context, r client.Client, namespace string) (string, error) {
	if v.Value != "" {
		return v.Value, nil
	} else if v.FromConfigMap != nil {
		secret := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: v.FromConfigMap.Name}, secret); err != nil {
			return "", err
		}
		return string(secret.Data[v.FromConfigMap.Key]), nil
	}
	return "", errors.New("cannot get private key, value and secret both empty")
}

func (v *ValueOrSecret) Get(ctx context.Context, r client.Client, namespace string) (string, error) {
	if v.Value != "" {
		return v.Value, nil
	} else if v.FromSecret != nil {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: v.FromSecret.Name}, secret); err != nil {
			return "", err
		}
		return string(secret.Data[v.FromSecret.Key]), nil
	}
	return "", errors.New("cannot get private key, value and secret both empty")
}
