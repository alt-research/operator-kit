// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

// util to get substrate peer id from nodekey

package subnodekey

import (
	"crypto/rand"

	"github.com/alt-research/operator-kit/hextool"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"golang.org/x/crypto/ed25519"
)

func GenerateNodeKey() []byte {
	seed := make([]byte, ed25519.SeedSize)
	_, _ = rand.Read(seed)
	return seed
}

// NodeKeyToPeerID gets the libp2p peerID string from substate node key, by using the nodekey as ed25519 seed
func NodeKeyToPeerID(nodekey string) (string, error) {
	keydata, err := hextool.Decode(nodekey)
	if err != nil {
		return "", err
	}
	pkey := ed25519.NewKeyFromSeed(keydata)
	if err != nil {
		return "", err
	}
	priv, err := crypto.UnmarshalEd25519PrivateKey([]byte(pkey))
	if err != nil {
		return "", err
	}
	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
