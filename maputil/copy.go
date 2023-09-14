// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package maputil

func Copy[T comparable, P any](dst *map[T]P, src map[T]P) {
	*dst = make(map[T]P, len(src))
	if len(src) == 0 {
		return
	}
	for k, v := range src {
		(*dst)[k] = v
	}
}

func Merge[T comparable, P any](dst *map[T]P, src map[T]P) {
	if len(src) == 0 {
		return
	}
	if dst == nil || *dst == nil {
		*dst = make(map[T]P)
	}
	for k, v := range src {
		if _, ok := (*dst)[k]; !ok {
			(*dst)[k] = v
		}
	}
}

func MergeOverwrite[T comparable, P any](dst *map[T]P, src map[T]P) {
	if len(src) == 0 {
		return
	}
	if dst == nil || *dst == nil {
		*dst = make(map[T]P)
	}
	for k, v := range src {
		(*dst)[k] = v
	}
}

func XMerge[T any](dst *map[string]T, src map[string]T) {
	if len(src) == 0 {
		return
	}
	if dst == nil || *dst == nil {
		*dst = make(map[string]T)
	}
	for k, v := range src {
		if _, ok := (*dst)[k]; ok && k != "" && k[0] == '-' {
			delete(*dst, k[1:])
		} else {
			(*dst)[k] = v
		}
	}
}
