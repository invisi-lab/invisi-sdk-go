package core

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// ProtocolContext identifies a chain and deployed protocol contract.
type ProtocolContext struct{ ChainID, ProtocolAddress fr.Element }

// Note is a private ordinary note. Its receiving address participates in its nullifier.
type Note struct{ NoteSecret, ReceivingAddress, TokenAddress, Amount, Nonce fr.Element }

// CreditNote uses an independent nonzero shared key; identity secrets stay private.
type CreditNote struct {
	SpenderIdentityCommitment, OwnerIdentityCommitment   fr.Element
	TokenAddress, Amount, Nonce, Expiry, CreditSharedKey fr.Element
}

// CreditRemainderRecovery contains circuit-bound fields from an authenticated draw event.
type CreditRemainderRecovery struct{ CreditNullifier, RemainderCommitment, RecoveryAmount fr.Element }
