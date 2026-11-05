package core

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// ComputeNoteIdentity randomizes the receiving owner address for one note.
func ComputeNoteIdentity(ownerAddress, nonce fr.Element) (fr.Element, error) {
	if ownerAddress.IsZero() || ElementToBigInt(ownerAddress).BitLen() > 160 {
		return fr.Element{}, fmt.Errorf("owner address must be nonzero")
	}
	return Poseidon2Hash(ElementFromUint64(DIdent), ownerAddress, nonce)
}

func ComputeNoteCommitment(chainID, identity, tokenAddress, amount fr.Element) (fr.Element, error) {
	for _, v := range []struct {
		x    fr.Element
		bits int
		name string
	}{{chainID, 64, "chainId"}, {tokenAddress, 160, "tokenAddress"}, {amount, 128, "amount"}} {
		if err := checkUint(v.x, v.bits, v.name); err != nil {
			return fr.Element{}, err
		}
	}
	if amount.IsZero() {
		return fr.Element{}, fmt.Errorf("amount must be nonzero")
	}
	return Poseidon2Hash(ElementFromUint64(DNote), chainID, identity, tokenAddress, amount)
}

// ComputeNoteNullifier binds the private note nonce and complete input leaf.
func ComputeNoteNullifier(chainID, nonce, noteCommitment fr.Element) (fr.Element, error) {
	if err := checkUint(chainID, 64, "chainId"); err != nil {
		return fr.Element{}, err
	}
	return Poseidon2Hash(ElementFromUint64(DNull), chainID, nonce, noteCommitment)
}
