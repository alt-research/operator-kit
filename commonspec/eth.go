// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package commonspec

import (
	"math/big"

	"github.com/alt-research/operator-kit/ethutil"
	"github.com/ethereum/go-ethereum/common"
)

// +kubebuilder:validation:Pattern:=`^(0x)?[0-9a-fA-F]{40}$`
type Address string

func (a Address) String() string {
	return string(a)
}

func (a Address) EthClientAddr() common.Address {
	return common.HexToAddress(a.String())
}

type AddressSlice []Address

func (x AddressSlice) Len() int           { return len(x) }
func (x AddressSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x AddressSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type AmountEth uint64

func (a AmountEth) Wei() *big.Int {
	return ethutil.Ether(uint64(a))
}
