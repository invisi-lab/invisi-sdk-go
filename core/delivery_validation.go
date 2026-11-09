package core

import (
	"fmt"
)

func validateNote(note Note) error {
	if err := ValidatePayload(note.Payload); err != nil {
		return err
	}
	if note.OwnerAddress.IsZero() || note.Amount.IsZero() {
		return fmt.Errorf("owner address and amount must be nonzero")
	}
	if err := checkUint(note.Amount, 128, "amount"); err != nil {
		return err
	}
	return checkUint(note.TokenAddress, 160, "tokenAddress")
}

func verifyNote(context NoteEncryptionContext, note Note) error {
	if err := validateProtocolContext(context.ProtocolContext, note.TokenAddress); err != nil {
		return err
	}
	if note.ProtocolContext != context.ProtocolContext {
		return fmt.Errorf("note context mismatch")
	}
	if err := validateNote(note); err != nil {
		return err
	}
	identity, err := ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	if err != nil {
		return err
	}
	leaf, err := ComputeNoteCommitment(context.ChainID, context.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	if err != nil {
		return err
	}
	if leaf != context.OutputCommitment {
		return fmt.Errorf("output commitment mismatch")
	}
	return nil
}
