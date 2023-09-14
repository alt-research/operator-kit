// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package tplutil

import "github.com/spf13/cast"

func incr(i interface{}) int64 {
	return cast.ToInt64(i) + 1
}

func init() {
	FuncMap["incr"] = incr
}
