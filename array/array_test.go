// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package array

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray(t *testing.T) {
	a1 := []int{1, 2, 3, 4, 5}
	assert.True(t, Contains(a1, 2))
	assert.False(t, Contains(a1, -1))
	Remove(&a1, 3)
	assert.EqualValues(t, []int{1, 2, 4, 5}, a1)
	a1 = []int{1, 2, 3, 3, 3, 5}
	Removes(&a1, 3, 2)
	assert.EqualValues(t, []int{1, 2, 3, 5}, a1)
	a2 := []int{}
	Removes(&a2, 3, 2)
}

// Slice([]int{1,2,3,4,5}, "1:3") -> []int{2,3}
// Slice([]int{1,2,3,4,5}, "1:") -> []int{2,3,4,5}
// Slice([]int{1,2,3,4,5}, ":-1") -> []int{1,2,3,4}
// Slice([]int{1,2,3,4,5}, "-2:") -> []int{4,5}
func TestSlice(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	assert.EqualValues(t, []int{2, 3}, Slice(a, "1:3"))
	assert.EqualValues(t, []int{2, 3, 4, 5}, Slice(a, "1:999"))
	assert.EqualValues(t, []int{2, 3, 4, 5}, Slice(a, "1:"))
	assert.EqualValues(t, []int{1, 2, 3, 4}, Slice(a, ":-1"))
	assert.EqualValues(t, []int{4, 5}, Slice(a, "-2:"))
}
