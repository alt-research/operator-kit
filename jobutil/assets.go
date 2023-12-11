// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package jobutil

import (
	"embed"
)

//go:embed scripts/*.sh
var assets embed.FS
