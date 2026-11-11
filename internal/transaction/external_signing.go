package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
)

type TransactionProveOptions struct {
	Prover Prover
	prover.Options
	Execution string
}

func ClonePayload(data []byte) []byte {
	if data == nil {
		return nil
	}
	return append([]byte{}, data...)
}

func (o *Execution) WalletRequest(ctx context.Context) (primitives.WalletAuthorizationRequest, error) {
	if err := o.CheckLease(ctx); err != nil {
		return primitives.WalletAuthorizationRequest{}, err
	}
	if !o.Plan.NeedsAuthorization || (o.Status != "prepared" && o.Status != "authorized") {
		return primitives.WalletAuthorizationRequest{}, errors.New("operation is not awaiting wallet authorization")
	}

	words, err := o.Words()
	if err != nil {
		return primitives.WalletAuthorizationRequest{}, err
	}
	return protocol.WalletAuthorizationRequest(o.Plan.Context, primitives.WalletOperation(o.Plan.Kind), o.Plan.Capacity, words[:len(words)-2])
}

func (o *Execution) ExportWalletAuthorization(ctx context.Context) (primitives.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	request, err := o.WalletRequest(ctx)
	if err != nil {
		return primitives.SigningRequest{}, err
	}
	typed, err := primitives.WalletAuthorizationTypedData(request)
	if err != nil {
		return primitives.SigningRequest{}, err
	}
	return primitives.NewSigningRequest(typed)
}

func (o *Execution) AttachWallet(ctx context.Context, signature []byte) error {
	request, err := o.WalletRequest(ctx)
	if err != nil {
		return err
	}
	witness, err := primitives.BuildWalletAuthorizationWitnessFromSignature(primitives.BigIntToElement(o.Runtime.Wallet.ownerAddress.Big()), request, append([]byte(nil), signature...))
	if err != nil {
		return err
	}
	o.Authorization = &witness
	o.Status = "authorized"
	return o.Persist(ctx, o.Status)
}

func (p *PreparedTransaction) ExportWalletAuthorization(ctx context.Context) (primitives.SigningRequest, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return primitives.SigningRequest{}, err
	}
	return p.operation.ExportWalletAuthorization(ctx)
}

func (p *PreparedTransaction) WithWalletSignature(ctx context.Context, signature []byte) (*PreparedTransaction, error) {
	copied := append([]byte(nil), signature...)
	return p.stage(ctx, func(o *Execution) error { return o.AttachWallet(ctx, copied) })
}

// Funding requests are reconstructed from the Frozen operation and current chain nonce.
func (o *Execution) FundingRequest(ctx context.Context) (apitypes.TypedData, error) {
	if err := o.CheckLease(ctx); err != nil {
		return apitypes.TypedData{}, err
	}
	if o.Plan.FundingMode != "signature" {
		return apitypes.TypedData{}, errors.New("operation does not require funding authorization")
	}
	if (o.Plan.Kind == "shield" && o.Status != "prepared") || (o.Plan.Kind != "shield" && (o.Status != "proved" || o.Bundle == nil)) {
		return apitypes.TypedData{}, errors.New("proof must precede funding authorization")
	}
	head, err := o.Runtime.Options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if head == nil || head.Time > o.Plan.Deadline || (o.Plan.Funding.Deadline != nil && new(big.Int).SetUint64(head.Time).Cmp(o.Plan.Funding.Deadline) > 0) {
		return apitypes.TypedData{}, errors.New("funding authorization expired")
	}
	caller, err := contracts.NewInvisiProtocolCaller(o.Plan.Context.ProtocolAddress, o.Runtime.Options.Provider)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	nonce, err := caller.Nonces(&bind.CallOpts{Context: ctx}, o.Plan.Funding.Payer)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if o.Plan.Funding.SigNonce == nil || nonce.Cmp(o.Plan.Funding.SigNonce) != 0 {
		return apitypes.TypedData{}, errors.New("prepared funding nonce is stale")
	}
	return o.FundingTypedData()
}

func (o *Execution) ExportFundingAuthorization(ctx context.Context) (primitives.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	plan, err := Frozen(o.Plan)
	if err != nil {
		return primitives.SigningRequest{}, err
	}
	copy := &Execution{Runtime: o.Runtime, Plan: plan, Status: o.Status, Bundle: o.Bundle, WorkerID: o.WorkerID}
	typed, err := copy.FundingRequest(ctx)
	if err != nil {
		return primitives.SigningRequest{}, err
	}
	request, err := primitives.NewSigningRequest(typed)
	if err != nil {
		return request, err
	}
	if err = o.CheckLease(ctx); err != nil {
		return primitives.SigningRequest{}, err
	}
	return request, nil
}

func (o *Execution) AttachFunding(ctx context.Context, signature []byte) error {
	typed, err := o.FundingRequest(ctx)
	if err != nil {
		return err
	}
	verifier := o.Runtime.Options.FundingVerifier
	if verifier == nil {
		verifier = GethFundingVerifier{Provider: o.Runtime.Options.Provider}
	}
	valid, err := verifier.VerifyTypedData(ctx, o.Plan.Funding.Payer, typed, signature)
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
		if owner, ok := RecoverTypedDataSigner(digest, signature); ok && owner == o.Plan.Funding.Payer {
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
		if owner, ok := RecoverTypedDataSigner(digest, signature); ok && owner == o.Plan.Funding.Payer {
			signature = append([]byte(nil), signature...)
			signature[64] += 27
		}
	}
	// Re-read nonce and deadline after potentially remote ERC-1271 verification.
	rechecked, err := o.FundingRequest(ctx)
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
	o.Plan.Funding.Signature = append([]byte(nil), signature...)
	if o.Plan.Target != "noir-recursive" {
		if err = o.Encode(); err != nil {
			return err
		}
	}
	o.Status = "ready"
	return o.Persist(ctx, o.Status)
}

func (p *PreparedTransaction) ExportFundingAuthorization(ctx context.Context) (primitives.SigningRequest, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return primitives.SigningRequest{}, err
	}
	return p.operation.ExportFundingAuthorization(ctx)
}

func (p *PreparedTransaction) WithFundingSignature(ctx context.Context, signature []byte) (*PreparedTransaction, error) {
	copied := append([]byte(nil), signature...)
	return p.stage(ctx, func(o *Execution) error { return o.AttachFunding(ctx, copied) })
}
