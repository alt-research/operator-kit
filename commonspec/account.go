// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package commonspec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type AccountPub struct {
	PublicKey string `json:"publicKey"`
	Address   string `json:"address"`
}

type AccountPubSpec struct {
	Name               string     `json:"name"`
	Index              int        `json:"index"`
	Eth                AccountPub `json:"eth,omitempty"`
	Aura               AccountPub `json:"aura,omitempty"`
	Grandpa            AccountPub `json:"grandpa,omitempty"`
	NodeKey            AccountPub `json:"nodeKey,omitempty"`
	ImOnline           AccountPub `json:"imOnline,omitempty"`
	AuthorityDiscovery AccountPub `json:"authorityDiscovery,omitempty"`
}

type AccountRef struct {
	//+optional
	Accounts []string `json:"accounts,omitempty"`
	//+optional
	Addresses []Address `json:"addresses,omitempty"`
	//+optional
	AccountSet string `json:"accountSet,omitempty"`
	//+optional
	//+kubebuilder:validation:Pattern="^([0-9]+(-[0-9]+)?)(,([0-9]+(-[0-9]+)?))*$"
	Selector string `json:"selector,omitempty"`
}

func (r *AccountRef) IsEmpty() bool {
	return len(r.Addresses) == 0 && len(r.Accounts) == 0 && r.AccountSet == ""
}

func (r *AccountRef) Validate() error {
	if lo.Count([]bool{len(r.Addresses) > 0, len(r.Accounts) > 0, r.AccountSet != ""}, true) > 1 {
		return errors.New("only one of address, name or accountSet can be specified")
	}
	return nil
}

func SelectorToIndexes(selector string) ([]int, error) {
	if selector == "" {
		return nil, nil
	}
	indexes := []int{}
	splited := strings.Split(selector, ",")
	for _, s := range splited {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.Contains(s, "-") {
			// range
			splited2 := strings.Split(s, "-")
			if len(splited2) != 2 {
				return nil, fmt.Errorf("invalid range: %s", s)
			}
			start, err := strconv.Atoi(splited2[0])
			if err != nil {
				return nil, fmt.Errorf("invalid range: %s", s)
			}
			end, err := strconv.Atoi(splited2[1])
			if err != nil {
				return nil, fmt.Errorf("invalid range: %s", s)
			}
			if start > end {
				return nil, fmt.Errorf("invalid range: %s", s)
			}
			for i := start; i <= end; i++ {
				indexes = append(indexes, i)
			}
		} else {
			// single index
			index, err := strconv.Atoi(s)
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", s)
			}
			indexes = append(indexes, index)
		}
	}
	return indexes, nil
}
