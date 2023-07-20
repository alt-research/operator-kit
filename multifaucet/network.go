package multifaucet

import "errors"

// Network registered in multifaucet
type Network struct {
	ChainID               string   `json:"chainId"`
	FaucetContractAddress string   `json:"faucetContractAddress"`
	RPC                   string   `json:"rpc"`
	Operator              string   `json:"operator,omitempty"`
	OperatorPrivateKey    string   `json:"operatorPrivateKey,omitempty"`
	SupportNative         bool     `json:"supportNative"`
	ERC20Tokens           []string `json:"erc20Tokens"`
}

type Result struct {
	Result string `json:"result"`
	Err    string `json:"error"`
}

func (r Result) Error() string {
	return r.Err
}

var ErrNotFound = errors.New("Not Found")
