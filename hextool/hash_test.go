package hextool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlake2(t *testing.T) {
	assert.Equal(t, "0x928b20366943e2afd11ebc0eae2e53a93bf177a4fcf35bcc64d503704e65e202", Blake2b256([]byte("test")))
}
