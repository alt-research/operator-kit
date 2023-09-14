// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package envs

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func AsMap(envs []string) map[string]string {
	m := make(map[string]string, len(envs))
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func AsEnvVars(envs []string) []corev1.EnvVar {
	vars := make([]corev1.EnvVar, len(envs))
	for i, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			vars[i] = corev1.EnvVar{Name: parts[0], Value: parts[1]}
		}
	}
	return vars
}
