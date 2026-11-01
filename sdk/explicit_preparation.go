package sdk

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func frozen[T any](value T) (T, error) {
	var result T
	raw, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(raw, &result)
	clear(raw)
	return result, err
}

func (i *Invisi) explicit(ctx context.Context, identity TransactionIdentity, funds ExplicitFunds, credits []CreditInput) (*WalletSession, error) {
	// Never retain or consult a configured Store in this mode.
	options := i.options
	options.Store = nil
	session, err := NewInvisi(options).resolve(ctx, ConnectionOptions{Wallet: identity.Wallet, Protocol: identity.Protocol})
	if err != nil {
		return nil, err
	}
	state := emptyState()
	for _, n := range funds.Inputs {
		if _, exists := state.Notes[n.Commitment]; exists {
			return nil, errors.New("duplicate note input")
		}
		if n.Note.ProtocolContext != session.coreContext() || n.Note.WalletIdentity != core.BigIntToElement(identity.Wallet.identity.Big()) {
			return nil, errors.New("note deployment or owner mismatch")
		}
		nf, err := identity.Wallet.ExportNullifier().ForNote(n.Note)
		if err != nil {
			return nil, err
		}
		state.Notes[n.Commitment] = StoredNote{Note: n.Note, Commitment: n.Commitment, Nullifier: nf, TreeID: n.TreeID, LeafIndex: n.LeafIndex}
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
		leaf, err := creditCommitment(c.State.Note)
		if err != nil || leaf != c.Commitment {
			return nil, errors.New("Credit commitment mismatch")
		}
		nf, err := core.ComputeCreditNullifier(c.State.Note.ChainID, c.State.Note.CreditSharedKey, core.BigIntToElement(leaf.Big()))
		if err != nil {
			return nil, err
		}
		state.Credits[creditKey(c.Commitment, c.Role)] = StoredCredit{State: c.State, Commitment: c.Commitment, Nullifier: fieldHash(nf), TreeID: c.TreeID, LeafIndex: c.LeafIndex, Role: c.Role}
	}
	session.explicitInventory = &state
	session.changeAddress = funds.ChangeAddress
	return session, nil
}

func (i *Invisi) PrepareTransfer(ctx context.Context, input ExplicitTransferPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = wrapPreparation("prepareTransfer", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, input.ExplicitFunds, nil)
	if err != nil {
		return nil, err
	}
	return detached(session.PrepareTransfer(ctx, input.TransferPreparation))
}

func (i *Invisi) PrepareShield(ctx context.Context, input ExplicitShieldPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = wrapPreparation("prepareShield", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, ExplicitFunds{}, nil)
	if err != nil {
		return nil, err
	}
	return detached(session.PrepareShield(ctx, input.ShieldPreparation))
}

func (i *Invisi) PrepareCreateCredit(ctx context.Context, input ExplicitCreateCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = wrapPreparation("prepareCreateCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, input.ExplicitFunds, nil)
	if err != nil {
		return nil, err
	}
	return detached(session.PrepareCreateCredit(ctx, input.CreateCreditPreparation))
}

func (i *Invisi) PrepareDrawCredit(ctx context.Context, input ExplicitDrawCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = wrapPreparation("prepareDrawCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, ExplicitFunds{}, input.CreditInputs)
	if err != nil {
		return nil, err
	}
	return detached(session.PrepareDrawCredit(ctx, input.DrawCreditPreparation))
}

func (i *Invisi) PrepareTopUpCredit(ctx context.Context, input ExplicitTopUpCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = wrapPreparation("prepareTopUpCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, input.ExplicitFunds, input.CreditInputs)
	if err != nil {
		return nil, err
	}
	return detached(session.PrepareTopUpCredit(ctx, input.TopUpCreditPreparation))
}

func (i *Invisi) PrepareCancelCredit(ctx context.Context, input ExplicitCancelCreditPreparation) (result *PreparedTransaction, failure error) {
	defer func() { failure = wrapPreparation("prepareCancelCredit", failure) }()
	identity := input.TransactionIdentity
	input.TransactionIdentity = TransactionIdentity{}
	input, err := frozen(input)
	if err != nil {
		return nil, err
	}
	session, err := i.explicit(ctx, identity, ExplicitFunds{}, input.CreditInputs)
	if err != nil {
		return nil, err
	}
	return detached(session.PrepareCancelCredit(ctx, input.CancelCreditPreparation))
}
