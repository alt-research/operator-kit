// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package subnodekey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testkey = "5c66ffd591ae3b6a57fd5a2015a3e23400aefea2d33f7e55e5fa138f18fedfae"
const testPeerId = "12D3KooWHjyYnisJ72B8DjZQTnCn2tm8KNu9gf8R9eNVJwtXK2Xf"

func TestNodeKey(t *testing.T) {
	id, err := NodeKeyToPeerID(testkey)
	assert.NoError(t, err)
	assert.Equal(t, testPeerId, id)
}
