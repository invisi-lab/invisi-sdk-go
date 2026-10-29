package core

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// ProtocolContext identifies a chain and deployed protocol contract.
type ProtocolContext struct{ ChainID, ProtocolAddress fr.Element }

// Note contains the private preimage of a deployment-scoped funds leaf, without spending keys.
type Note struct {
	ProtocolContext
	WalletIdentity, TokenAddress, Amount, Nonce fr.Element
	// Memo is optional authenticated delivery metadata, excluded from commitments.
	Memo []byte
}

// CreditNote binds both wallet roles, public payout addresses and the shared recovery secret.
type CreditNote struct {
	ProtocolContext
	SpenderWalletIdentity, OwnerWalletIdentity, SpenderPayout, OwnerPayout fr.Element
	TokenAddress, Amount, Nonce, Expiry, CreditSharedKey                   fr.Element
	// Memo is optional authenticated delivery metadata, excluded from commitments.
	Memo []byte
}

// CreditRemainderRecovery contains circuit-bound fields from an authenticated draw event.
type CreditRemainderRecovery struct{ CreditNullifier, RemainderCommitment, RecoveryAmount fr.Element }

// CreditDeliveryState retains both receiving key snapshots for autonomous lifecycle recovery.
type CreditDeliveryState struct {
	Note                                     CreditNote
	OwnerReceivingKeys, SpenderReceivingKeys ReceivingKeys
}
