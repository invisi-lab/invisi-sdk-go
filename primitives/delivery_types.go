package primitives

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

const ordinaryNoteDomain = "Invisi.NoteDelivery.v1"

type NoteEncryptionContext struct {
	ProtocolContext
	OutputCommitment fr.Element
}

type NoteDecryptionContext struct {
	NoteEncryptionContext
	TokenAddress, OwnerAddress fr.Element
	Clue                       []byte
}
