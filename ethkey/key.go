// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package ethkey

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Key struct {
	ecdsa.PrivateKey
}

func (k *Key) Address() common.Address {
	return crypto.PubkeyToAddress(k.PublicKey)
}
