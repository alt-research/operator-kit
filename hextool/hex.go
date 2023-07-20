package hextool

import (
	"github.com/alt-research/operator-kit/must"
	ethhex "github.com/ethereum/go-ethereum/common/hexutil"
)

var Encode = ethhex.Encode

// Has0xPrefix validates str begins with '0x' or '0X'.
func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// Decode decodes a hex string.
func Decode(input string) ([]byte, error) {
	if Has0xPrefix(input) {
		return ethhex.Decode(input)
	}
	return ethhex.Decode("0x" + input)
}

// MustDecode decodes a hex string. It panics for invalid input.
func MustDecode(input string) []byte {
	return must.Two(Decode(input))
}
