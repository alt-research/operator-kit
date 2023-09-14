// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package must

import (
	"reflect"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Two[T any](val T, err error) T {
	Must(err)
	return val
}

func Three[T any, P any](val1 T, val2 P, err error) (T, P) {
	Must(err)
	return val1, val2
}

func Four[T any, P any, Q any](val1 T, val2 P, val3 Q, err error) (T, P, Q) {
	Must(err)
	return val1, val2, val3
}

// Default retruns the first non-empty value
func Default[T any](values ...T) T {
	var value T
	for _, value = range values {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice:
			if !v.IsNil() && v.Len() > 0 {
				return value
			}
		default:
			if !v.IsZero() {
				return value
			}
		}
	}
	return value
}
