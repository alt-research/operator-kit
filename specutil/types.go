// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package specutil

import (
	corev1 "k8s.io/api/core/v1"
)

type ValueOrConfigMap struct {
	Value         string                       `json:"value,omitempty"`
	FromConfigMap *corev1.ConfigMapKeySelector `json:"fromConfigMap,omitempty"`
}

type ValueOrSecret struct {
	Value      string                    `json:"value,omitempty"`
	FromSecret *corev1.SecretKeySelector `json:"fromSecret,omitempty"`
}
