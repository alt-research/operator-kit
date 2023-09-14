// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package fp

// Ident returns the value passed in.
func Ident[T any](v T) T {
	return v
}
