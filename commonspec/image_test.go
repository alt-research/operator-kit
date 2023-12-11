// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package commonspec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraints(t *testing.T) {
	assert.False(t, *ImageRef{Tag: "v0.6.0-xxx-v1"}.VersionMatch("< 0.7.0"))
	assert.True(t, *ImageRef{Tag: "v0.6.0"}.VersionMatch("< 0.7.0"))
	assert.True(t, *ImageRef{Tag: "v0.6.0"}.VersionMatch("< 0.7.0-pre"))
	assert.True(t, *ImageRef{Tag: "v0.7.1"}.VersionMatch("< 0.7.2-pre"))
	assert.False(t, *ImageRef{Tag: "v0.7.2"}.VersionMatch("< 0.7.2-pre"))
	assert.True(t, *ImageRef{Tag: "v0.6.0-xxx-v1"}.VersionMatch("< 0.7.0-pre"))
	assert.True(t, *ImageRef{Tag: "v0.6.0-xxx-v1"}.VersionMatch("0.6.x-pre"))
	assert.False(t, *ImageRef{Tag: "v0.6.0-xxx-v1"}.VersionMatch("0.7.x-pre"))
	assert.True(t, *ImageRef{Tag: "v0.6.0"}.VersionMatch("0.6.x-pre"))
	assert.False(t, *ImageRef{Tag: "v0.6.0"}.VersionMatch("0.7.x-pre"))
	assert.True(t, *ImageRef{Tag: "v0.6.0-xxx-v1"}.VersionMatch("^0.6-pre"))
	assert.False(t, *ImageRef{Tag: "v0.6.0-xxx-v1"}.VersionMatch("^0.7-pre"))
	assert.True(t, *ImageRef{Tag: "v0.6.0"}.VersionMatch("^0.6-pre"))
	assert.False(t, *ImageRef{Tag: "v0.6.0"}.VersionMatch("^0.7-pre"))
}
