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
	leaf, err := ComputeNoteCommitment(context.ChainID, identity, note.TokenAddress, note.Amount)
	if err != nil {
		return err
	}
	if leaf != context.OutputCommitment {
		return fmt.Errorf("output commitment mismatch")
	}
	return nil
}

func verifyCredit(context NoteEncryptionContext, note CreditNote) error {
	if err := validateProtocolContext(context.ProtocolContext, note.TokenAddress); err != nil {
		return err
	}
	if note.ProtocolContext != context.ProtocolContext {
		return fmt.Errorf("Credit context mismatch")
	}
	if err := validateCredit(note); err != nil {
		return err
	}
	keyHash, err := ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if err != nil {
		return err
	}
	leaf, err := ComputeCreditCommitment(context.ChainID, note.SpenderWalletIdentity, note.OwnerWalletIdentity, note.SpenderPayout, note.OwnerPayout, note.TokenAddress, note.Amount, note.Nonce, note.Expiry, keyHash, note.CreditSpendingState)
	if err != nil {
		return err
	}
	if leaf != context.OutputCommitment {
		return fmt.Errorf("credit commitment mismatch")
	}
	return nil
}
