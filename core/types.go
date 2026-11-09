package core

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// ProtocolContext identifies a chain and deployed protocol contract.
type ProtocolContext struct{ ChainID, ProtocolAddress fr.Element }

// Note contains the private preimage of a deployment-scoped funds leaf, without spending keys.
type Note struct {
	ProtocolContext
	OwnerAddress, TokenAddress, Amount, Nonce fr.Element
	// Payload is opaque authenticated application data, excluded from commitments.
	Payload []byte
}
