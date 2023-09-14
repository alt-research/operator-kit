// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package str

import "strings"

func EqualFold[T, P ~string](a T, b P) bool {
	return strings.EqualFold(string(a), string(b))
}
