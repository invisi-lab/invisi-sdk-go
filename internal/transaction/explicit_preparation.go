package transaction

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
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

func (i *Invisi) explicit(ctx context.Context, identity TransactionIdentity, funds ExplicitFunds) (*Runtime, error) {
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
		if n.Note.ProtocolContext != session.CoreContext() || n.Note.OwnerAddress != primitives.BigIntToElement(identity.Wallet.identity.Big()) {
			return nil, errors.New("note deployment or owner mismatch")
		}
		nf, err := identity.Wallet.ExportNullifier().ForNote(n.Note)
		if err != nil {
			return nil, err
		}
		state.Notes[n.Commitment] = NoteRecord{Note: n.Note, Commitment: n.Commitment, Nullifier: nf, TreeID: n.TreeID, LeafIndex: n.LeafIndex}
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
	session, err := i.explicit(ctx, identity, input.ExplicitFunds)
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
	session, err := i.explicit(ctx, identity, ExplicitFunds{})
	if err != nil {
		return nil, err
	}
	return Detached(session.PrepareShield(ctx, input.ShieldPreparation))
}
