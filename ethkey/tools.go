package ethkey

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/alt-research/operator-kit/hextool"
	"github.com/alt-research/operator-kit/must"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

const (
	AddressLength        = 20
	PrivateKeyLength     = 64
	PublickeyLength      = 65
	CompressPubkeyLength = 33
)

func GenerateKey() *ecdsa.PrivateKey {
	return must.Two(crypto.GenerateKey())
}

func NewMnemonic12() string {
	return must.Two(bip39.NewMnemonic(must.Two(bip39.NewEntropy(128))))
}

func NewMnemonic24() string {
	return must.Two(bip39.NewMnemonic(must.Two(bip39.NewEntropy(256))))
}

func FromMnemonic(mnemonic string, path string) (*ecdsa.PrivateKey, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	dPath, err := ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}
	return dPath.Derive(masterKey)
}

func PrivateKeyToAddress(k *ecdsa.PrivateKey) string {
	return crypto.PubkeyToAddress(k.PublicKey).Hex()
}

func HexToPri(p string) (*ecdsa.PrivateKey, error) {
	d, err := hextool.Decode(p)
	if err != nil {
		return nil, err
	}
	k, err := crypto.ToECDSA(d)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func HexToPub(p string) (k *ecdsa.PublicKey, err error) {
	d, err := hextool.Decode(p)
	if err != nil {
		return nil, err
	}
	switch len(d) {
	case PublickeyLength:
		k, err = crypto.UnmarshalPubkey(d)
	case CompressPubkeyLength:
		k, err = crypto.DecompressPubkey(d)
	default:
		return nil, fmt.Errorf("invalid public key")
	}
	return
}

func PriToHex(k *ecdsa.PrivateKey) string {
	return hextool.Encode(crypto.FromECDSA(k))
}

func PubToShortHex(k *ecdsa.PublicKey) string {
	return hextool.Encode(crypto.CompressPubkey(k))
}

func PubToHex(k *ecdsa.PublicKey) string {
	return hextool.Encode(crypto.FromECDSAPub(k))
}

func PubToAddress(k *ecdsa.PublicKey) string {
	return crypto.PubkeyToAddress(*k).Hex()
}

func HexToAddress(a string) (*common.Address, error) {
	d, err := hextool.Decode(a)
	if err != nil {
		return nil, err
	}
	if len(d) != AddressLength {
		return nil, fmt.Errorf("invalid address")
	}
	addr := common.BytesToAddress(d)
	return &addr, nil
}

func ChecksumAddressHex(a string) (string, error) {
	addr, err := HexToAddress(a)
	if err != nil {
		return "", err
	}
	return addr.Hex(), nil
}
