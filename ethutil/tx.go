package ethutil

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Transfer(ctx context.Context, client *ethclient.Client, chainID *big.Int, from *ecdsa.PrivateKey, to common.Address, amount *big.Int, timeout time.Duration) (*types.Transaction, *types.Receipt, error) {
	log := log.FromContext(ctx)
	if from == nil {
		return nil, nil, errors.New("from account is required")
	}
	fromAddress := crypto.PubkeyToAddress(from.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, nil, err
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, nil, err
	}
	tx := types.NewTx(&types.LegacyTx{
		To:       &to,
		Value:    amount,
		Nonce:    nonce,
		Gas:      30000,
		GasPrice: gasPrice,
	})
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		return signedTx, nil, err
	}
	log.Info("sending transaction", "tx", signedTx.Hash().Hex(), "from", fromAddress.Hex(), "to", to.Hex(), "amount", amount.String())
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		if !strings.Contains(err.Error(), "already known") {
			return signedTx, nil, errors.Wrapf(err, "failed to send transaction %s from %s to %s with amount %s", signedTx.Hash().Hex(), fromAddress.Hex(), to.Hex(), amount.String())
		}
	}
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	receipt, err := WaitTx(c, client, signedTx.Hash())
	return signedTx, receipt, errors.Wrapf(err, "failed to make transfer %s from %s to %s with amount %s", signedTx.Hash().Hex(), fromAddress.Hex(), to.Hex(), amount.String())
}

var (
	TransactionFailed   = errors.New("Transaction failed.")
	TransactionNotFound = errors.New("Transaction not found.")
)

func WaitTx(ctx context.Context, client *ethclient.Client, tx common.Hash, opts ...retry.Option) (receipt *types.Receipt, err error) {
	err = retry.Do(
		func() error {
			receipt, err = client.TransactionReceipt(ctx, tx)
			if err != nil {
				return err
			}
			if receipt.Status == types.ReceiptStatusFailed {
				return errors.Wrapf(TransactionFailed, "transaction %s failed", tx)
			}
			return nil
		},
		retry.Context(ctx),
		retry.Attempts(100),
		retry.Delay(1*time.Second),
		retry.RetryIf(func(err error) bool {
			return !errors.Is(err, TransactionFailed) && !errors.Is(err, TransactionNotFound)
		}),
	)
	return receipt, errors.Wrapf(err, "Fail to wait transaction (tx=%s)", tx)
}
