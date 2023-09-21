package subkey

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ecdsa"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

// +kubebuilder:validation:Enum:=ecdsa;sr25519;ed25519
type KeyScheme string

const (
	KeySchemeEcdsa   KeyScheme = "ecdsa"
	KeySchemeSr25519 KeyScheme = "sr25519"
	KeySchemeEd25519 KeyScheme = "ed25519"
)

var schemeMap = map[KeyScheme]subkey.Scheme{
	KeySchemeEcdsa:   &ecdsa.Scheme{},
	KeySchemeSr25519: &sr25519.Scheme{},
	KeySchemeEd25519: &ed25519.Scheme{},
}

var ErrInvalidScheme = fmt.Errorf("invalid scheme")

// +kubebuilder:validation:Enum:=babe;gran;acco;imon;audi;stak;dumm
type KeyType string

const (
	KeyTypeAccount            KeyType = "acco"
	KeyTypeAura               KeyType = "aura"
	KeyTypeAuthorityDiscovery KeyType = "audi"
	KeyTypeBabe               KeyType = "babe"
	KeyTypeDummy              KeyType = "dumm"
	KeyTypeGrandpa            KeyType = "gran"
	KeyTypeImOnline           KeyType = "imon"
	KeyTypeStaking            KeyType = "stak"
)

type KeyPair struct {
	subkey.KeyPair
	Scheme     KeyScheme
	Type       KeyType
	SS58Format uint16
}

var DefaultSS58Format uint16 = 42

func DeriveKeyPair(suri string, scheme KeyScheme) (*KeyPair, error) {
	if s, ok := schemeMap[scheme]; !ok {
		return nil, ErrInvalidScheme
	} else {
		kp, err := subkey.DeriveKeyPair(s, suri)
		if err != nil {
			return nil, err
		}
		return &KeyPair{
			KeyPair:    kp,
			Scheme:     scheme,
			SS58Format: DefaultSS58Format,
		}, nil
	}
}

func (k *KeyPair) SetType(typ KeyType) {
	k.Type = typ
}

func (k *KeyPair) SetSS58Format(network uint16) {
	k.SS58Format = network
}

func (k *KeyPair) PrivateKey() string {
	return hexutil.Encode(k.KeyPair.Seed())
}

func (k *KeyPair) PublicKey() string {
	return hexutil.Encode(k.KeyPair.Public())
}

func (k *KeyPair) KeystoreFilename(typ ...KeyType) string {
	if len(typ) > 0 {
		return hexutil.Encode([]byte(typ[0]))[2:] + k.PublicKey()[2:]
	}
	if k.Type != "" {
		return hexutil.Encode([]byte(k.Type))[2:] + k.PublicKey()[2:]
	}
	return ""
}

func (k *KeyPair) SS58Address() string {
	return k.KeyPair.SS58Address(k.SS58Format)
}
