// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package maputil

func Pick[T comparable, P any](src map[T]P, picker func(T, P) bool) map[T]P {
	dst := make(map[T]P)
	if len(src) == 0 {
		return dst
	}
	for k, v := range src {
		if picker(k, v) {
			dst[k] = v
		}
	}
	return dst
}
