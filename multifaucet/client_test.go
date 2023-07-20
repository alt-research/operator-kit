package multifaucet

import (
	"context"
	"testing"
)

func TestClient(t *testing.T) {
	t.Skip()
	c := NewClient("https://stg-faucet.alt.technology", "rAv4pfH9OYqwczoyIoFQ2CMIVr6FFjld")
	// t.Log(c.GetAll(context.Background()))
	// t.Log(c.Get(context.Background(), "9993"))
	// t.Log(c.Get(context.Background(), "9990"))
	t.Log(c.Upsert(context.Background(), &Network{
		ChainID:               "9998",
		FaucetContractAddress: "0xD740F70415e0B16f5026728Cd26E8aE2473c003b",
		OperatorPrivateKey:    "0x94c49300a58d576011786bcb006aa06f5a91b34b4383891e8029c21dc39fbb8b",
		RPC:                   "https://testnet-beacon-api.altlayer.io",
		SupportNative:         true,
	}))
	t.Log(c.Delete(context.Background(), "9998"))
}
