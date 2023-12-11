// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package subwasm

import (
	"os/exec"
)

func Exec(args []string) ([]byte, error) {
	return exec.Command("subwasm", args...).Output()
}

func Info(file string) ([]byte, error) {
	return Exec([]string{"info", "--json", file})
}
