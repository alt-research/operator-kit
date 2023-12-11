// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package hextool

import (
	"crypto/sha256"
	"os"

	"golang.org/x/crypto/blake2b"
)

func SHA256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return Encode(h.Sum(nil))
}

func SHA256OfFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return SHA256(data), nil
}

func Blake2b256(data []byte) string {
	h, _ := blake2b.New256(nil)
	h.Write(data)
	return Encode(h.Sum(nil))
}

func Blake2b256OfFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Blake2b256(data), nil
}
