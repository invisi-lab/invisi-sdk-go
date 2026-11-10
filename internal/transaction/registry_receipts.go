package transaction

import (
	"bytes"
	"context"
	"errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (o *RegistrationOperation) trackInternal(ctx context.Context, hash common.Hash) (*RegistrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if hash == (common.Hash{}) || (o.status != "ready" && o.status != "uncertain") {
		return nil, errors.New("ready registration and transaction hash required")
	}
	o.transactionHash, o.status = hash, "uncertain"
	o.receiptBlock = nil
	return o, nil
}

func (o *RegistrationOperation) reconcileInternal(ctx context.Context) (Confirmation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	result := Confirmation{o.status, o.transactionHash}
	if o.transactionHash == (common.Hash{}) {
		return result, errors.New("registration is not submitted")
	}
	uncertain := func() (Confirmation, error) {
		o.status = "uncertain"
		o.receiptBlock = nil
		return Confirmation{o.status, o.transactionHash}, nil
	}
	receipt, err := o.registry.provider.TransactionReceipt(ctx, o.transactionHash)
	if errors.Is(err, ethereum.NotFound) {
		return uncertain()
	}
	if err != nil {
		return result, err
	}
	if receipt == nil {
		return uncertain()
	}
	if receipt.TxHash != o.transactionHash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		return result, errors.New("registration receipt transaction or block mismatch")
	}
	var receiptHeader *types.Header
	canonical := func() (bool, error) {
		header, err := o.registry.provider.HeaderByNumber(ctx, receipt.BlockNumber)
		if err != nil {
			return false, err
		}
		valid := header != nil && header.Number != nil && header.Number.Cmp(receipt.BlockNumber) == 0 && header.Hash() == receipt.BlockHash
		if valid {
			receiptHeader = types.CopyHeader(header)
		}
		return valid, nil
	}
	valid, err := canonical()
	if err != nil {
		return result, err
	}
	if !valid {
		return uncertain()
	}
	reader, ok := o.registry.provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return result, errors.New("provider must authenticate registration calldata")
	}
	transaction, pending, err := reader.TransactionByHash(ctx, o.transactionHash)
	if err != nil {
		return result, err
	}
	if pending || transaction == nil || transaction.Hash() != o.transactionHash || transaction.To() == nil || *transaction.To() != o.registry.deployment.Address || !bytes.Equal(transaction.Data(), o.data) || transaction.Value().Sign() != 0 {
		return result, errors.New("registration transaction mismatch")
	}
	if o.direct {
		sender, err := types.Sender(types.LatestSignerForChainID(o.context.ChainID), transaction)
		if err != nil || sender != o.wallet.identity {
			return result, errors.New("direct registration sender differs from owner")
		}
	}
	status := "failed"
	if receipt.Status == types.ReceiptStatusSuccessful {
		if err = o.registry.ValidateDeployment(ctx); err != nil {
			return result, err
		}
		if err = o.verifyReceivingKeysEvent(receipt); err != nil {
			return result, err
		}
		status = "confirmed"
	}
	// RPC calls above can cross a reorganization. Recheck the receipt anchor
	// after authenticating calldata and the transaction's own event.
	valid, err = canonical()
	if err != nil {
		return result, err
	}
	if !valid {
		return uncertain()
	}
	o.status = status
	o.receiptBlock = &protocolrpc.CanonicalBlock{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash, ParentHash: receiptHeader.ParentHash, Timestamp: receiptHeader.Time}
	return Confirmation{o.status, o.transactionHash}, nil
}

func (o *RegistrationOperation) verifyReceivingKeysEvent(receipt *types.Receipt) error {
	parsed, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		return err
	}
	filterer, err := contracts.NewInvisiAddressRegistryFilterer(o.registry.deployment.Address, nil)
	if err != nil {
		return err
	}
	count := 0
	eventID := parsed.Events["ReceivingKeysSet"].ID
	for _, log := range receipt.Logs {
		if log == nil || log.Address != o.registry.deployment.Address || len(log.Topics) == 0 || log.Topics[0] != eventID {
			continue
		}
		if log.Removed || log.BlockHash != receipt.BlockHash || log.BlockNumber != receipt.BlockNumber.Uint64() || log.TxHash != receipt.TxHash || log.TxIndex != receipt.TransactionIndex || len(log.Topics) != 2 {
			return errors.New("registration event transaction or block mismatch")
		}
		event, err := filterer.ParseReceivingKeysSet(*log)
		if err != nil {
			return err
		}
		if event.WalletIdentity != o.registration.OwnerAddress || !bytes.Equal(event.ViewingPublicKey, o.registration.Keys.ViewingPublicKey) || event.ClueKey != o.registration.Keys.ClueKey || event.NextNonce != o.registration.Nonce+1 {
			return errors.New("registration event differs from authorized keys")
		}
		count++
	}
	if count != 1 {
		return errors.New("exactly one authorized receiving-key event required")
	}
	return nil
}
