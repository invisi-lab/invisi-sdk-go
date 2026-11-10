package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
)

func (o *RegistrationOperation) Preview() RegistrationPreview {
	o.mu.Lock()
	defer o.mu.Unlock()
	preview := RegistrationPreview{RegistryAddress: o.registry.deployment.Address, OwnerAddress: o.wallet.identity, Update: o.options.Update, Nonce: o.registration.Nonce, Deadline: o.registration.Deadline, Status: o.status, ReceivingKeys: primitives.ReceivingKeys{ViewingPublicKey: append([]byte(nil), o.registration.Keys.ViewingPublicKey...), ClueKey: append([]byte(nil), o.registration.Keys.ClueKey[:]...)}}
	if o.receiptBlock != nil {
		block := *o.receiptBlock
		preview.ReceiptBlock = &block
	}
	return preview
}

func (o *RegistrationOperation) ExportTransaction(ctx context.Context) (TransactionCall, error) {
	if err := ctx.Err(); err != nil {
		return TransactionCall{}, err
	}
	o.lifecycle.mu.Lock()
	defer o.lifecycle.mu.Unlock()
	if o.lifecycle.busy || o.lifecycle.generation != o.generation {
		return TransactionCall{}, errors.New("registration handle is busy or stale")
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.data) == 0 || o.status == "prepared" {
		return TransactionCall{}, errors.New("authorized registration required")
	}
	if _, err := o.currentContext(ctx); err != nil {
		return TransactionCall{}, err
	}
	return TransactionCall{To: o.registry.deployment.Address, Data: append([]byte(nil), o.data...), Value: new(big.Int)}, nil
}

// currentContext rejects a stale signature before signing or exporting it.
func (o *RegistrationOperation) currentContext(ctx context.Context) (protocol.RegistryContext, error) {
	current := o.context
	if err := o.registry.ValidateDeployment(ctx); err != nil {
		return current, err
	}
	head, err := o.registry.provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return current, err
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return current, errors.New("latest block unavailable")
	}
	if head.Time > o.registration.Deadline {
		return current, errors.New("registration expired")
	}
	caller, err := contracts.NewInvisiAddressRegistryCaller(o.registry.deployment.Address, o.registry.provider)
	if err != nil {
		return current, err
	}
	nonce, err := caller.NonceOf(&bind.CallOpts{Context: ctx, BlockNumber: head.Number}, o.registration.OwnerAddress)
	if err != nil {
		return current, err
	}
	if nonce != o.registration.Nonce {
		return current, errors.New("registry nonce changed; prepare again")
	}
	canonical, err := o.registry.provider.HeaderByNumber(ctx, head.Number)
	if err != nil {
		return current, err
	}
	if canonical == nil || canonical.Number == nil || canonical.Number.Cmp(head.Number) != 0 || canonical.Hash() != head.Hash() {
		return current, errors.New("registration read block is not canonical")
	}
	current.Now = new(big.Int).SetUint64(head.Time)
	return current, nil
}
