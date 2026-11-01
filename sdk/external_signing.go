package sdk

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type TransactionProveOptions struct {
	Prover Prover
	prover.Options
	Execution string
}

func clonePayload(data []byte) []byte {
	if data == nil {
		return nil
	}
	return append([]byte{}, data...)
}

func (o *Operation) walletRequest(ctx context.Context) (core.WalletAuthorizationRequest, error) {
	if err := o.checkLease(ctx); err != nil {
		return core.WalletAuthorizationRequest{}, err
	}
	if !o.plan.NeedsAuthorization || (o.status != "prepared" && o.status != "authorized") {
		return core.WalletAuthorizationRequest{}, errors.New("operation is not awaiting wallet authorization")
	}
	if err := o.checkDrawPeriod(ctx); err != nil {
		return core.WalletAuthorizationRequest{}, err
	}
	words, err := o.words()
	if err != nil {
		return core.WalletAuthorizationRequest{}, err
	}
	return protocol.WalletAuthorizationRequest(o.plan.Context, core.WalletOperation(o.plan.Kind), o.plan.Capacity, words[:len(words)-2])
}
func (o *Operation) ExportWalletAuthorization(ctx context.Context) (core.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	request, err := o.walletRequest(ctx)
	if err != nil {
		return core.SigningRequest{}, err
	}
	typed, err := core.WalletAuthorizationTypedData(request)
	if err != nil {
		return core.SigningRequest{}, err
	}
	return core.NewSigningRequest(typed)
}
func (o *Operation) attachWallet(ctx context.Context, signature []byte) error {
	request, err := o.walletRequest(ctx)
	if err != nil {
		return err
	}
	witness, err := core.BuildWalletAuthorizationWitnessFromSignature(core.BigIntToElement(o.session.wallet.identity.Big()), request, append([]byte(nil), signature...))
	if err != nil {
		return err
	}
	o.authorization = &witness
	o.status = "authorized"
	return o.persist(ctx, o.status)
}
func (o *Operation) WithWalletSignature(ctx context.Context, signature []byte) (*Operation, error) {
	copied := append([]byte(nil), signature...)
	return o.stage(ctx, func(next *Operation) error { return next.attachWallet(ctx, copied) })
}
func (p *PreparedTransaction) ExportWalletAuthorization(ctx context.Context) (core.SigningRequest, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return core.SigningRequest{}, err
	}
	return p.operation.ExportWalletAuthorization(ctx)
}
func (p *PreparedTransaction) WithWalletSignature(ctx context.Context, signature []byte) (*PreparedTransaction, error) {
	copied := append([]byte(nil), signature...)
	return p.stage(ctx, func(o *Operation) error { return o.attachWallet(ctx, copied) })
}

// Funding requests are reconstructed from the frozen operation and current chain nonce.
func (o *Operation) fundingRequest(ctx context.Context) (apitypes.TypedData, error) {
	if err := o.checkLease(ctx); err != nil {
		return apitypes.TypedData{}, err
	}
	if o.plan.FundingMode != "signature" {
		return apitypes.TypedData{}, errors.New("operation does not require funding authorization")
	}
	if (o.plan.Kind == "shield" && o.status != "prepared") || (o.plan.Kind != "shield" && (o.status != "proved" || o.bundle == nil)) {
		return apitypes.TypedData{}, errors.New("proof must precede funding authorization")
	}
	head, err := o.session.options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if head == nil || head.Time > o.plan.Deadline || (o.plan.Funding.Deadline != nil && new(big.Int).SetUint64(head.Time).Cmp(o.plan.Funding.Deadline) > 0) {
		return apitypes.TypedData{}, errors.New("funding authorization expired")
	}
	caller, err := contracts.NewInvisiProtocolCaller(o.plan.Context.ProtocolAddress, o.session.options.Provider)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	nonce, err := caller.Nonces(&bind.CallOpts{Context: ctx}, o.plan.Funding.Payer)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if o.plan.Funding.SigNonce == nil || nonce.Cmp(o.plan.Funding.SigNonce) != 0 {
		return apitypes.TypedData{}, errors.New("prepared funding nonce is stale")
	}
	return o.fundingTypedData()
}
func (o *Operation) ExportFundingAuthorization(ctx context.Context) (core.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	plan, err := frozen(o.plan)
	if err != nil {
		return core.SigningRequest{}, err
	}
	copy := &Operation{session: o.session, plan: plan, status: o.status, bundle: o.bundle, workerID: o.workerID}
	typed, err := copy.fundingRequest(ctx)
	if err != nil {
		return core.SigningRequest{}, err
	}
	request, err := core.NewSigningRequest(typed)
	if err != nil {
		return request, err
	}
	if err = o.checkLease(ctx); err != nil {
		return core.SigningRequest{}, err
	}
	return request, nil
}
func (o *Operation) attachFunding(ctx context.Context, signature []byte) error {
	typed, err := o.fundingRequest(ctx)
	if err != nil {
		return err
	}
	verifier := o.session.options.FundingVerifier
	if verifier == nil {
		verifier = GethFundingVerifier{Provider: o.session.options.Provider}
	}
	valid, err := verifier.VerifyTypedData(ctx, o.plan.Funding.Payer, typed, signature)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid funding signature")
	}
	if len(signature) == 64 {
		digest, _, err := apitypes.TypedDataAndHash(typed)
		if err != nil {
			return err
		}
		if owner, ok := recoverTypedDataSigner(digest, signature); ok && owner == o.plan.Funding.Payer {
			expanded := make([]byte, 65)
			copy(expanded, signature)
			expanded[64] = expanded[32] >> 7
			expanded[32] &= 0x7f
			signature = expanded
		}
	}
	if len(signature) == 65 && signature[64] <= 1 {
		digest, _, err := apitypes.TypedDataAndHash(typed)
		if err != nil {
			return err
		}
		if owner, ok := recoverTypedDataSigner(digest, signature); ok && owner == o.plan.Funding.Payer {
			signature = append([]byte(nil), signature...)
			signature[64] += 27
		}
	}
	// Re-read nonce and deadline after potentially remote ERC-1271 verification.
	rechecked, err := o.fundingRequest(ctx)
	if err != nil {
		return err
	}
	before, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return err
	}
	after, _, err := apitypes.TypedDataAndHash(rechecked)
	if err != nil {
		return err
	}
	if common.BytesToHash(before) != common.BytesToHash(after) {
		return errors.New("funding nonce changed")
	}
	o.plan.Funding.Signature = append([]byte(nil), signature...)
	if o.plan.Target != "noir-recursive" {
		if err = o.encode(); err != nil {
			return err
		}
	}
	o.status = "ready"
	return o.persist(ctx, o.status)
}
func (o *Operation) WithFundingSignature(ctx context.Context, signature []byte) (*Operation, error) {
	copied := append([]byte(nil), signature...)
	return o.stage(ctx, func(next *Operation) error { return next.attachFunding(ctx, copied) })
}
func (p *PreparedTransaction) ExportFundingAuthorization(ctx context.Context) (core.SigningRequest, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return core.SigningRequest{}, err
	}
	return p.operation.ExportFundingAuthorization(ctx)
}
func (p *PreparedTransaction) WithFundingSignature(ctx context.Context, signature []byte) (*PreparedTransaction, error) {
	copied := append([]byte(nil), signature...)
	return p.stage(ctx, func(o *Operation) error { return o.attachFunding(ctx, copied) })
}
