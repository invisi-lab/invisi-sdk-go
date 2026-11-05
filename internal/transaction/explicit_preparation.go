package transaction

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func Frozen[T any](value T) (T, error) {
	var result T
	raw, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(raw, &result)
	clear(raw)
	return result, err
}

func (i *Invisi) explicit(ctx context.Context, identity TransactionIdentity, funds ExplicitFunds, credits []CreditInput) (*Runtime, error) {
	// Never retain or consult a configured Store in this mode.
	options := i.options
	session, err := ResolveRuntime(ctx, options, identity.Wallet, identity.Protocol, "explicit")
	if err != nil {
		return nil, err
	}
	state := EmptyInventory()
	for _, n := range funds.Inputs {
		if _, exists := state.Notes[n.Commitment]; exists {
			return nil, errors.New("duplicate note input")
		}
		if n.Note.ProtocolContext != session.CoreContext() || n.Note.OwnerAddress != core.BigIntToElement(identity.Wallet.identity.Big()) {
			return nil, errors.New("note deployment or owner mismatch")
		}
		nf, err := identity.Wallet.ExportNullifier().ForNote(n.Note)
		if err != nil {
			return nil, err
		}
		state.Notes[n.Commitment] = NoteRecord{Note: n.Note, Commitment: n.Commitment, Nullifier: nf, TreeID: n.TreeID, LeafIndex: n.LeafIndex}
	}
	seen := map[common.Hash]bool{}
	for _, c := range credits {
		if c.Role != "owner" && c.Role != "spender" {
			return nil, errors.New("invalid Credit input role")
		}
		if seen[c.Commitment] {
			return nil, errors.New("duplicate Credit input")
		}
		seen[c.Commitment] = true
		leaf, err := CreditCommitment(c.State.Note)
		if err != nil || leaf != c.Commitment {
			return nil, errors.New("Credit commitment mismatch")
		}
		nf, err := core.ComputeCreditNullifier(c.State.Note.ChainID, c.State.Note.CreditSharedKey, core.BigIntToElement(leaf.Big()))
		if err != nil {
			return nil, err
		}
		state.Credits[CreditKey(c.Commitment, c.Role)] = CreditRecord{State: c.State, Commitment: c.Commitment, Nullifier: FieldHash(nf), TreeID: c.TreeID, LeafIndex: c.LeafIndex, Role: c.Role}
	}
	session.ExplicitInventory = &state
	session.ChangeAddress = funds.ChangeAddress
	return session, nil
}

func (i *Invisi) PrepareTransfer(ctx context.Context, input ExplicitTransferPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = WrapPreparation("prepareTransfer", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := Frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, input.ExplicitFunds, nil)
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareTransfer(ctx, input.TransferPreparation))
}

func (i *Invisi) PrepareShield(ctx context.Context, input ExplicitShieldPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = WrapPreparation("prepareShield", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := Frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, ExplicitFunds{}, nil)
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareShield(ctx, input.ShieldPreparation))
}

func (i *Invisi) PrepareCreateCredit(ctx context.Context, input ExplicitCreateCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = WrapPreparation("prepareCreateCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := Frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, input.ExplicitFunds, nil)
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareCreateCredit(ctx, input.CreateCreditPreparation))
}

func (i *Invisi) PrepareDrawCredit(ctx context.Context, input ExplicitDrawCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = WrapPreparation("prepareDrawCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := Frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, ExplicitFunds{}, input.CreditInputs)
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareDrawCredit(ctx, input.DrawCreditPreparation))
}

func (i *Invisi) PrepareTopUpCredit(ctx context.Context, input ExplicitTopUpCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = WrapPreparation("prepareTopUpCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := Frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, input.ExplicitFunds, input.CreditInputs)
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareTopUpCredit(ctx, input.TopUpCreditPreparation))
}

func (i *Invisi) PrepareCancelCredit(ctx context.Context, input ExplicitCancelCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = WrapPreparation("prepareCancelCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := Frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, ExplicitFunds{}, input.CreditInputs)
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareCancelCredit(ctx, input.CancelCreditPreparation))
}
