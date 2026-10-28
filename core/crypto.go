package core

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// DeriveNoteSecret implements the optional offchain invoice identity policy.
// Credit shared keys must instead be independently sampled with GenerateCreditSharedKey.
func DeriveNoteSecret(masterKey, invoiceID fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DNoteSecret), masterKey, invoiceID)
}

func ComputeIdentityCommitment(noteSecret, receivingAddress fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DIdent), noteSecret, receivingAddress)
}

func ComputeNoteCommitment(chainID, identityCommitment, tokenAddress, amount, nonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DNote), chainID, identityCommitment, tokenAddress, amount, nonce)
}

// ComputeChangeCommitment creates an ordinary note using caller-supplied secrets and nonce.
func ComputeChangeCommitment(chainID, changeNoteSecret, changeReceivingAddress, tokenAddress, amount, nonce fr.Element) (fr.Element, error) {
	identity, err := ComputeIdentityCommitment(changeNoteSecret, changeReceivingAddress)
	if err != nil {
		return fr.Element{}, err
	}
	return ComputeNoteCommitment(chainID, identity, tokenAddress, amount, nonce)
}

// ComputeNoteNullifier binds both the private note secret and receiving address.
func ComputeNoteNullifier(chainID, noteSecret, receivingAddress, tokenAddress, amount, nonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DNull), chainID, noteSecret, receivingAddress, tokenAddress, amount, nonce)
}

func ComputeCreditSharedKeyHash(creditSharedKey fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCreditKeyHash), creditSharedKey)
}

func ComputeCreditCommitment(chainID, spenderIdentityCommitment, ownerIdentityCommitment, tokenAddress, amount, nonce, expiry, creditSharedKeyHash fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCredit), chainID, spenderIdentityCommitment, ownerIdentityCommitment, tokenAddress, amount, nonce, expiry, creditSharedKeyHash)
}

func ComputeCreditNullifier(chainID, creditSharedKey, tokenAddress, amount, nonce, expiry fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCNull), chainID, creditSharedKey, tokenAddress, amount, nonce, expiry)
}

// DeriveRemainderNonce scopes a Credit remainder nonce to the deployment and consumed note.
func DeriveRemainderNonce(chainID, protocolAddress, tokenAddress, creditNullifier, creditSharedKey fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCreditRemainderNonce), chainID, protocolAddress, tokenAddress, creditNullifier, creditSharedKey)
}

// ComputeCreditRecoveryPad derives the additive pad used by Credit draw recovery.
func ComputeCreditRecoveryPad(chainID, protocolAddress, tokenAddress, creditNullifier, creditSharedKey fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DCreditRecovery), chainID, protocolAddress, tokenAddress, creditNullifier, creditSharedKey)
}
