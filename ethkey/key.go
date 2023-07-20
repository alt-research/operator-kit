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
