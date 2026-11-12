package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"math/big"
	"testing"
)

type creditReceiptProvider struct {
	sessionProvider
	transaction *types.Transaction
	receipt     *types.Receipt
}

func (p creditReceiptProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return p.receipt, nil
}
func (p creditReceiptProvider) TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error) {
	return p.transaction, false, nil
}
func TestCreditReconciliationRequiresExactCanonicalTransaction(t *testing.T) {
	ctx := context.Background()
	wallet, token, _ := testSession(t)
	base := wallet.options.Provider.(sessionProvider)
	data := []byte{1, 2, 3, 4}
	to := wallet.scope.ProtocolAddress
	tx := types.NewTx(&types.DynamicFeeTx{ChainID: big.NewInt(31337), To: &to, Data: data})
	receipt := &types.Receipt{TxHash: tx.Hash(), BlockNumber: base.head.Number, BlockHash: base.head.Hash(), Status: types.ReceiptStatusFailed}
	wallet.options.Provider = creditReceiptProvider{base, tx, receipt}
	intent := credit.Intent{Version: 2, Context: credit.Context{ProtocolContext: wallet.coreContext(), TokenAddress: token, Timestamp: base.head.Time}, Revoke: &credit.RevokeParams{Token: token, Owner: wallet.wallet.OwnerAddress(), Shard: 1, Deadline: big.NewInt(400)}}
	reconciler := wallet.CreditReconciler("latest")
	outcome, err := reconciler.Reconcile(ctx, intent, data, tx.Hash())
	if err != nil || outcome == nil || outcome.Success {
		t.Fatal("canonical reversion not authenticated", err)
	}
	if _, err = reconciler.Reconcile(ctx, intent, []byte{9}, tx.Hash()); err == nil {
		t.Fatal("accepted another transaction's calldata")
	}
	receipt.BlockHash = common.HexToHash("0xdead")
	if _, err = reconciler.Reconcile(ctx, intent, data, tx.Hash()); err == nil {
		t.Fatal("accepted noncanonical receipt")
	}
}
