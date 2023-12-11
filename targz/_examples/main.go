// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alt-research/operator-kit/targz"
)

func main() {
	// Create a temporary file structiure we can use
	tmpDir, dirToCompress := createExampleData()

	// Comress a folder to my_archive.tar.gz
	err := targz.Compress(dirToCompress, filepath.Join(tmpDir, "my_archive.tar.gz"))
	if err != nil {
		fmt.Println("Comress error")
		panic(err)
		os.Exit(1)
	}

	// Extract my_archive.tar.gz to a new folder called extracted
	err = targz.Extract(filepath.Join(tmpDir, "my_archive.tar.gz"), filepath.Join(tmpDir, "extracted"))
	if err != nil {
		fmt.Println("Extract error")
		panic(err)
		os.Exit(1)
	}

	// Open so we can se the files and remove the directory if we'd like.
	cmd := exec.Command("open", tmpDir)
	cmd.Run()

	os.Exit(0)
}

func createExampleData() (string, string) {
	tmpDir, err := os.MkdirTemp("", "targz-example-*")
	if err != nil {
		fmt.Println("tmpdir error")
		panic(err)
		os.Exit(1)
	}

	directory := filepath.Join(tmpDir, "my_folder")
	subDirectory := filepath.Join(directory, "my_sub_folder")
	err = os.MkdirAll(subDirectory, 0o755)
	if err != nil {
		fmt.Println("mkdir error")
		panic(err)
		os.Exit(1)
	}

	_, err = os.Create(filepath.Join(subDirectory, "my_file.txt"))
	if err != nil {
		fmt.Println("create file error")
		panic(err)
		os.Exit(1)
	}

	return tmpDir, directory
}
