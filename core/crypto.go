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
func ComputeCreditSharedKeyHash(key fr.Element) (fr.Element, error) {
	if key.IsZero() {
		return fr.Element{}, fmt.Errorf("Credit shared key must be nonzero")
	}
	return Poseidon2Hash(ElementFromUint64(DCreditKeyHash), key)
}
func ComputeCreditCommitment(chainID, spenderWalletIdentity, ownerWalletIdentity, spenderPayout, ownerPayout, tokenAddress, amount, nonce, expiry, keyHash fr.Element, state CreditSpendingState) (fr.Element, error) {
	for _, v := range []struct {
		x    fr.Element
		bits int
		name string
	}{{chainID, 64, "chainId"}, {tokenAddress, 160, "tokenAddress"}, {spenderPayout, 160, "spenderPayout"}, {ownerPayout, 160, "ownerPayout"}, {amount, 126, "amount"}, {expiry, 64, "expiry"}} {
		if err := checkUint(v.x, v.bits, v.name); err != nil {
			return fr.Element{}, err
		}
	}
	if spenderWalletIdentity.IsZero() || ownerWalletIdentity.IsZero() || spenderPayout.IsZero() || ownerPayout.IsZero() || amount.IsZero() || expiry.IsZero() {
		return fr.Element{}, fmt.Errorf("Credit identities, payouts, amount and expiry must be positive")
	}
	if spenderWalletIdentity != spenderPayout || ownerWalletIdentity != ownerPayout {
		return fr.Element{}, fmt.Errorf("Credit payout must equal its owner identity")
	}
	if err := ValidateCreditSpendingState(state); err != nil {
		return fr.Element{}, err
	}
	fields := []fr.Element{ElementFromUint64(DCredit), chainID, spenderWalletIdentity, ownerWalletIdentity, spenderPayout, ownerPayout, tokenAddress, amount, nonce, expiry, keyHash}
	return Poseidon2Hash(append(fields, state.Fields()...)...)
}

// ComputeCreditNullifier binds the entire grant, preventing alternate-grant shared-key collisions.
func ComputeCreditNullifier(chainID, key, creditCommitment fr.Element) (fr.Element, error) {
	if err := checkUint(chainID, 64, "chainId"); err != nil {
		return fr.Element{}, err
	}
	if key.IsZero() {
		return fr.Element{}, fmt.Errorf("Credit shared key must be nonzero")
	}
	return Poseidon2Hash(ElementFromUint64(DCNull), chainID, key, creditCommitment)
}
func DeriveRemainderNonce(chainID, protocolAddress, tokenAddress, creditNullifier, key fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCreditRemainderNonce), chainID, protocolAddress, tokenAddress, creditNullifier, key)
}
func ComputeCreditRecoveryPad(chainID, protocolAddress, tokenAddress, creditNullifier, key fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCreditRecovery), chainID, protocolAddress, tokenAddress, creditNullifier, key)
}
