package core

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// DeriveNullifierKey derives a nullifier key from a master key and invoice ID.
// Formula: Poseidon2Hash(D_NKEY, masterKey, invoiceId)
func DeriveNullifierKey(masterKey, invoiceId fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainNKey), masterKey, invoiceId)
}

// DeriveCreditNullifierKey derives the credit shared secret from the receiver's
// master key and the credit ID. The server owns this derivation and delivers the
// result to the payer in the IPP 6.2 authorize response; this function exists for
// receiver-side and tooling use. Domain-separated from DeriveNullifierKey (tag 10
// vs 1) so neither key reveals the other.
// Formula: Poseidon2Hash(D_CREDIT_NKEY, receiverMasterKey, creditId)
func DeriveCreditNullifierKey(receiverMasterKey, creditId fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainCreditNKey), receiverMasterKey, creditId)
}

// ComputeIdentityCommitment computes the identity commitment.
// Formula: Poseidon2Hash(D_IDENT, nullifierKey, receivingAddress)
func ComputeIdentityCommitment(nullifierKey, receivingAddress fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainIdent), nullifierKey, receivingAddress)
}

// ComputePaymentCommitment computes the payment commitment.
// Formula: Poseidon2Hash(D_PAY, chainId, identityCommitment, tokenAddress, amount, nonce)
func ComputePaymentCommitment(chainId, identityCommitment, tokenAddress, amount, nonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainPay), chainId, identityCommitment, tokenAddress, amount, nonce)
}

// ComputeChangeCommitment computes a change commitment for shielded payments.
// First computes changeIdentityCommitment, then uses it in a payment commitment.
func ComputeChangeCommitment(chainId, changeNullifierKey, changeReceivingAddress, tokenAddress, amount, nonce fr.Element) (fr.Element, error) {
	changeIC, err := ComputeIdentityCommitment(changeNullifierKey, changeReceivingAddress)
	if err != nil {
		return fr.Element{}, err
	}
	return Poseidon2Hash(ElementFromUint64(DomainPay), chainId, changeIC, tokenAddress, amount, nonce)
}

// ComputePaymentNullifier computes the payment nullifier.
// Formula: Poseidon2Hash(D_NULL, chainId, nullifierKey, tokenAddress, amount, nonce)
func ComputePaymentNullifier(chainId, nullifierKey, tokenAddress, amount, nonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainNull), chainId, nullifierKey, tokenAddress, amount, nonce)
}

// ComputeNullifierKeyHash computes the hash of a nullifier key.
// Formula: Poseidon2Hash(D_NKHASH, nullifierKey)
func ComputeNullifierKeyHash(nullifierKey fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainNKHash), nullifierKey)
}

// ComputeCreditCommitment computes a credit commitment.
// Formula: Poseidon2Hash(D_CREDIT, chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash)
func ComputeCreditCommitment(chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainCredit), chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash)
}

// ComputeCreditNullifier computes a credit nullifier.
// Formula: Poseidon2Hash(D_CNULL, chainId, nullifierKey, tokenAddress, amount, nonce, expiry)
func ComputeCreditNullifier(chainId, nullifierKey, tokenAddress, amount, nonce, expiry fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainCNull), chainId, nullifierKey, tokenAddress, amount, nonce, expiry)
}

// DeriveRemainderNonce derives a new nonce for credit partial draw remainder.
// Formula: Poseidon2Hash(D_DERIVE, 0, nonce, newNonce)
func DeriveRemainderNonce(nonce, newNonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainDerive), ElementFromUint64(0), nonce, newNonce)
}

// DeriveRemainderNullifierKey derives a new nullifier key for credit partial draw remainder.
// Formula: Poseidon2Hash(D_DERIVE, 1, nullifierKey, newNonce)
func DeriveRemainderNullifierKey(nullifierKey, newNonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainDerive), ElementFromUint64(1), nullifierKey, newNonce)
}

// ComputeRefundLink computes a refund link for audit purposes.
// Formula: Poseidon2Hash(D_REFLINK, originalCommitment, refundNonce)
func ComputeRefundLink(originalCommitment, refundNonce fr.Element) (fr.Element, error) {
	return Poseidon2Hash(ElementFromUint64(DomainRefLink), originalCommitment, refundNonce)
}
