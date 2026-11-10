package transaction

import (
	"bytes"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func (o *Execution) ValidateRecovered() error {
	p := o.Plan
	if p.Version != 1 || p.OwnerAddress != o.Runtime.Wallet.identity || p.Namespace != o.Runtime.Scope.Namespace {
		return errors.New("recovered wallet scope mismatch")
	}
	if err := o.ValidateBroadcasterPlan(); err != nil {
		return err
	}
	for _, output := range p.Outputs {
		if err := Positive(output.Amount, 128); err != nil {
			return err
		}
		identity := primitives.BigIntToElement(output.Identity.Big())
		leaf, err := primitives.ComputeNoteCommitment(o.Runtime.CoreContext().ChainID, o.Runtime.CoreContext().ProtocolAddress, identity, primitives.BigIntToElement(p.Context.TokenAddress.Big()), primitives.BigIntToElement(output.Amount))
		if err != nil || FieldHash(leaf) != output.Commitment {
			return errors.New("recovered output commitment mismatch")
		}
		if output.Note != nil {
			note := output.Note
			if note.ProtocolContext != o.Runtime.CoreContext() || note.TokenAddress != primitives.BigIntToElement(p.Context.TokenAddress.Big()) || primitives.ElementToBigInt(note.Amount).Cmp(output.Amount) != 0 {
				return errors.New("recovered output scope mismatch")
			}
			expected, err := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
			if err != nil || FieldHash(expected) != output.Identity {
				return errors.New("recovered output opening mismatch")
			}
		}
		if (len(output.Delivery.Ciphertext) == 0) != (len(output.Delivery.Clue) == 0) {
			return errors.New("recovered output delivery mismatch")
		}
	}
	for _, input := range p.Inputs {
		note := input.Note
		if note.ProtocolContext != o.Runtime.CoreContext() || note.OwnerAddress != primitives.BigIntToElement(p.OwnerAddress.Big()) || note.TokenAddress != primitives.BigIntToElement(p.Context.TokenAddress.Big()) {
			return errors.New("recovered note scope mismatch")
		}
		identity, err := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
		if err != nil {
			return err
		}
		leaf, err := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
		if err != nil || FieldHash(leaf) != input.Commitment {
			return errors.New("recovered note commitment mismatch")
		}
		nf, err := o.Runtime.Wallet.ExportNullifier().ForNote(note)
		if err != nil || nf != input.Nullifier {
			return errors.New("recovered note nullifier mismatch")
		}
	}

	if p.Kind != "shield" {
		words, err := o.Words()
		if err != nil {
			return err
		}
		inputs := make([]common.Hash, len(words))
		for i := range words {
			inputs[i] = words[i]
		}
		statement, err := prover.Statement(o.CircuitName(), p.Target, inputs)
		if err != nil {
			return err
		}
		if o.Bundle != nil {
			if err = prover.AssertProofBundle(*o.Bundle, statement); err != nil {
				return err
			}
		}
	}
	if len(o.Data) > 0 {
		if p.Kind != "shield" && o.Bundle == nil {
			return errors.New("recovered calldata has no proof")
		}
		stored := append([]byte(nil), o.Data...)
		if err := o.Encode(); err != nil {
			return err
		}
		if !bytes.Equal(stored, o.Data) {
			return errors.New("recovered transaction differs from Frozen operation")
		}
	}
	return nil
}
