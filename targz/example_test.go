// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package targz_test

import (
	"os"
	"path/filepath"

	"github.com/alt-research/operator-kit/targz"
)

func Example() {
	// Create a temporary file structiure we can use
	tmpDir, dirToCompress := createExampleData()

	// Comress a folder to my_archive.tar.gz
	err := targz.Compress(dirToCompress, filepath.Join(tmpDir, "my_archive.tar.gz"))
	if err != nil {
		panic(err)
	}

	// Extract my_archive.tar.gz to a new folder called extracted
	err = targz.Extract(filepath.Join(tmpDir, "my_archive.tar.gz"), filepath.Join(tmpDir, "extracted"))
	if err != nil {
		panic(err)
	}
}

func createExampleData() (string, string) {
	tmpDir, err := os.MkdirTemp("", "targz-example-*")
	if err != nil {
		panic(err)
	}

	directory := filepath.Join(tmpDir, "my_folder")
	subDirectory := filepath.Join(directory, "my_sub_folder")
	err = os.MkdirAll(subDirectory, 0o755)
	if err != nil {
		panic(err)
	}

	_, err = os.Create(filepath.Join(subDirectory, "my_file.txt"))
	if err != nil {
		panic(err)
	}

	return tmpDir, directory
}
