// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package envs

import (
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func SliceToMap(envs []string) map[string]string {
	m := make(map[string]string, len(envs))
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func MapToEnvVars(envs map[string]string) []corev1.EnvVar {
	vars := make([]corev1.EnvVar, len(envs))
	i := 0
	for k, v := range envs {
		vars[i] = corev1.EnvVar{Name: k, Value: v}
		i++
	}
	return vars
}

func SliceToEnvVars(envs []string) []corev1.EnvVar {
	return MapToEnvVars(SliceToMap(envs))
}

func EnvVarsToSlice(envs []corev1.EnvVar) ([]string, error) {
	slice := make([]string, len(envs))
	for i, env := range envs {
		if env.ValueFrom != nil {
			return nil, errors.Errorf("env var %s has valueFrom, not supported", env.Name)
		}
		slice[i] = env.Name + "=" + env.Value
	}
	return slice, nil
}

func EnvVarsToMap(envs []corev1.EnvVar) (map[string]string, error) {
	m := make(map[string]string, len(envs))
	for _, env := range envs {
		if env.ValueFrom != nil {
			return nil, errors.Errorf("env var %s has valueFrom, not supported", env.Name)
		}
		m[env.Name] = env.Value
	}
	return m, nil
}
