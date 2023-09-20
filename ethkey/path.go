// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package ethkey

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/alt-research/operator-kit/must"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

func ParseDerivationPath(path string) (*DerivationPath, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path: derivation path empty")
	}

	// clean all the parts of any trim spaces
	for indx := range parts {
		parts[indx] = strings.TrimSpace(parts[indx])
	}

	// first part has to be an 'm'
	if parts[0] != "m" {
		return nil, fmt.Errorf("invalid path: first has to be m")
	}

	result := DerivationPath{}
	for _, p := range parts[1:] {
		val := new(big.Int)
		if strings.HasSuffix(p, "'") {
			p = strings.TrimSuffix(p, "'")
			val.Add(val, decVal)
		}

		bigVal, ok := new(big.Int).SetString(p, 0)
		if !ok {
			return nil, fmt.Errorf("invalid path: parts should be integers")
		}
		val.Add(val, bigVal)

		// TODO, limit to uint32
		if !val.IsUint64() {
			return nil, fmt.Errorf("invalid path: parts overflowed, should be uint32")
		}
		result = append(result, uint32(val.Uint64()))
	}

	return &result, nil
}

type DerivationPath []uint32

// 0x800000
var decVal = big.NewInt(2147483648)

const (
	DefaultDerivationPathPrefix = "m/44'/60'/0'/0/"
	DefaultPath                 = DefaultDerivationPathPrefix + "0"
)

func DefaultPathIndexed(n int) string {
	return fmt.Sprintf("%s%d", DefaultDerivationPathPrefix, n)
}

// DefaultDerivationPath is the default derivation path for Ethereum addresses
// var DefaultDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0}
var DefaultDerivationPath = *must.Two(ParseDerivationPath(DefaultPath))

func (d *DerivationPath) Derive(master *hdkeychain.ExtendedKey) (*ecdsa.PrivateKey, error) {
	var err error
	key := master
	for _, n := range *d {
		key, err = key.Derive(n)
		if err != nil {
			return nil, err
		}
	}
	priv, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return priv.ToECDSA(), nil
}
