package ethkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertKey(t *testing.T) {
	require.Equal(t, DefaultDerivationPath, DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0})

	priv, err := HexToPri("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2FF80")
	assert.NoError(t, err)
	assert.EqualValues(t, "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", PriToHex(priv))
	pub, err := HexToPub("0x038318535b54105d4a7aae60c08fc45f9687181b4fdfc625bd1a753fa7397FED75")
	assert.NoError(t, err)
	assert.EqualValues(t, "0x038318535b54105d4a7aae60c08fc45f9687181b4fdfc625bd1a753fa7397fed75", PubToShortHex(pub))
	assert.EqualValues(t, "0x038318535b54105d4a7aae60c08fc45f9687181b4fdfc625bd1a753fa7397fed75", PubToShortHex(&priv.PublicKey))
	addr, err := ChecksumAddressHex("f39fd6e51aad88f6f4ce6ab8827279cfffb92266")
	assert.NoError(t, err)
	assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", addr)
}
