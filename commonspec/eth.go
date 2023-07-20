package commonspec

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/umbracle/ethgo"
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
	return ethgo.Ether(uint64(a))
}
