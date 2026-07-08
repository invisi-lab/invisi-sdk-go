package core

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// testElements returns a set of deterministic test field elements.
func testElements(t *testing.T) []fr.Element {
	t.Helper()
	vals := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	elems := make([]fr.Element, len(vals))
	for i, v := range vals {
		elems[i] = ElementFromUint64(v)
	}
	return elems
}

// TestDeriveNullifierKey verifies the formula matches direct Poseidon2Hash call.
func TestDeriveNullifierKey(t *testing.T) {
	e := testElements(t)
	masterKey, invoiceId := e[0], e[1]

	got, err := DeriveNullifierKey(masterKey, invoiceId)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainNKey), masterKey, invoiceId)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("DeriveNullifierKey: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputeIdentityCommitment verifies the formula matches direct Poseidon2Hash call.
func TestComputeIdentityCommitment(t *testing.T) {
	e := testElements(t)
	nullifierKey, receivingAddress := e[0], e[1]

	got, err := ComputeIdentityCommitment(nullifierKey, receivingAddress)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainIdent), nullifierKey, receivingAddress)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputeIdentityCommitment: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputePaymentCommitment verifies the formula matches direct Poseidon2Hash call.
func TestComputePaymentCommitment(t *testing.T) {
	e := testElements(t)
	chainId, identityCommitment, tokenAddress, amount, nonce := e[0], e[1], e[2], e[3], e[4]

	got, err := ComputePaymentCommitment(chainId, identityCommitment, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainPay), chainId, identityCommitment, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputePaymentCommitment: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputeChangeCommitment verifies that it correctly computes identity commitment internally.
func TestComputeChangeCommitment(t *testing.T) {
	e := testElements(t)
	chainId, changeNullifierKey, changeReceivingAddress, tokenAddress, amount, nonce := e[0], e[1], e[2], e[3], e[4], e[5]

	got, err := ComputeChangeCommitment(chainId, changeNullifierKey, changeReceivingAddress, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	// Manually compute via ComputeIdentityCommitment then ComputePaymentCommitment.
	changeIC, err := ComputeIdentityCommitment(changeNullifierKey, changeReceivingAddress)
	if err != nil {
		t.Fatal(err)
	}
	want, err := Poseidon2Hash(ElementFromUint64(DomainPay), chainId, changeIC, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputeChangeCommitment: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputeChangeCommitmentDiffersFromPayment verifies change commitment differs
// from a payment commitment using the raw identity commitment (not re-derived).
func TestComputeChangeCommitmentDiffersFromPayment(t *testing.T) {
	e := testElements(t)
	chainId, nullifierKey, addr, tokenAddress, amount, nonce := e[0], e[1], e[2], e[3], e[4], e[5]

	change, err := ComputeChangeCommitment(chainId, nullifierKey, addr, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	// Using addr directly as identity commitment (different from re-deriving).
	payment, err := ComputePaymentCommitment(chainId, addr, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if change.Equal(&payment) {
		t.Error("ComputeChangeCommitment should differ from ComputePaymentCommitment with raw address as identity")
	}
}

// TestComputePaymentNullifier verifies the formula matches direct Poseidon2Hash call.
func TestComputePaymentNullifier(t *testing.T) {
	e := testElements(t)
	chainId, nullifierKey, tokenAddress, amount, nonce := e[0], e[1], e[2], e[3], e[4]

	got, err := ComputePaymentNullifier(chainId, nullifierKey, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainNull), chainId, nullifierKey, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputePaymentNullifier: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputeNullifierKeyHash verifies the formula matches direct Poseidon2Hash call.
func TestComputeNullifierKeyHash(t *testing.T) {
	e := testElements(t)
	nullifierKey := e[0]

	got, err := ComputeNullifierKeyHash(nullifierKey)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainNKHash), nullifierKey)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputeNullifierKeyHash: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputeCreditCommitment verifies the formula matches direct Poseidon2Hash call.
func TestComputeCreditCommitment(t *testing.T) {
	e := testElements(t)
	chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash :=
		e[0], e[1], e[2], e[3], e[4], e[5], e[6], e[7]

	got, err := ComputeCreditCommitment(chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainCredit), chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputeCreditCommitment: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestComputeCreditNullifier verifies the formula matches direct Poseidon2Hash call.
func TestComputeCreditNullifier(t *testing.T) {
	e := testElements(t)
	chainId, nullifierKey, tokenAddress, amount, nonce, expiry := e[0], e[1], e[2], e[3], e[4], e[5]

	got, err := ComputeCreditNullifier(chainId, nullifierKey, tokenAddress, amount, nonce, expiry)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainCNull), chainId, nullifierKey, tokenAddress, amount, nonce, expiry)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputeCreditNullifier: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestDeriveRemainderNonce verifies the formula matches direct Poseidon2Hash call.
func TestDeriveRemainderNonce(t *testing.T) {
	e := testElements(t)
	nonce, newNonce := e[0], e[1]

	got, err := DeriveRemainderNonce(nonce, newNonce)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainDerive), ElementFromUint64(0), nonce, newNonce)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("DeriveRemainderNonce: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestDeriveRemainderNullifierKey verifies the formula matches direct Poseidon2Hash call.
func TestDeriveRemainderNullifierKey(t *testing.T) {
	e := testElements(t)
	nullifierKey, newNonce := e[0], e[1]

	got, err := DeriveRemainderNullifierKey(nullifierKey, newNonce)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainDerive), ElementFromUint64(1), nullifierKey, newNonce)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("DeriveRemainderNullifierKey: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestRemainderFunctionsDistinguishable verifies DeriveRemainderNonce and
// DeriveRemainderNullifierKey produce different outputs for the same inputs
// (distinguisher 0 vs 1).
func TestRemainderFunctionsDistinguishable(t *testing.T) {
	e := testElements(t)
	input1, input2 := e[0], e[1]

	nonceDerived, err := DeriveRemainderNonce(input1, input2)
	if err != nil {
		t.Fatal(err)
	}

	keyDerived, err := DeriveRemainderNullifierKey(input1, input2)
	if err != nil {
		t.Fatal(err)
	}

	if nonceDerived.Equal(&keyDerived) {
		t.Error("DeriveRemainderNonce and DeriveRemainderNullifierKey must produce different outputs")
	}
}

// TestComputeRefundLink verifies the formula matches direct Poseidon2Hash call.
func TestComputeRefundLink(t *testing.T) {
	e := testElements(t)
	originalCommitment, refundNonce := e[0], e[1]

	got, err := ComputeRefundLink(originalCommitment, refundNonce)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainRefLink), originalCommitment, refundNonce)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("ComputeRefundLink: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestDomainIsolation verifies that different domain tags produce different outputs
// for the same inputs.
func TestDomainIsolation(t *testing.T) {
	e := testElements(t)
	a, b := e[0], e[1]

	// Use two functions that take the same number of inputs but have different domains.
	nkey, err := DeriveNullifierKey(a, b)
	if err != nil {
		t.Fatal(err)
	}

	ident, err := ComputeIdentityCommitment(a, b)
	if err != nil {
		t.Fatal(err)
	}

	if nkey.Equal(&ident) {
		t.Error("domain isolation failure: DeriveNullifierKey and ComputeIdentityCommitment returned same value")
	}

	// Commitment vs nullifier (same arity, different domain).
	chainId, nullifierKey, tokenAddress, amount, nonce := e[0], e[1], e[2], e[3], e[4]

	commitment, err := ComputePaymentCommitment(chainId, nullifierKey, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	nullifier, err := ComputePaymentNullifier(chainId, nullifierKey, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if commitment.Equal(&nullifier) {
		t.Error("domain isolation failure: ComputePaymentCommitment and ComputePaymentNullifier returned same value")
	}
}

// TestDeterminism verifies that calling the same function twice with the same
// inputs produces the same result.
func TestDeterminism(t *testing.T) {
	e := testElements(t)
	masterKey, invoiceId := e[0], e[1]

	r1, err := DeriveNullifierKey(masterKey, invoiceId)
	if err != nil {
		t.Fatal(err)
	}

	r2, err := DeriveNullifierKey(masterKey, invoiceId)
	if err != nil {
		t.Fatal(err)
	}

	if !r1.Equal(&r2) {
		t.Error("DeriveNullifierKey is not deterministic")
	}

	// Also check ComputePaymentCommitment determinism.
	chainId, identityCommitment, tokenAddress, amount, nonce := e[0], e[1], e[2], e[3], e[4]

	c1, err := ComputePaymentCommitment(chainId, identityCommitment, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	c2, err := ComputePaymentCommitment(chainId, identityCommitment, tokenAddress, amount, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if !c1.Equal(&c2) {
		t.Error("ComputePaymentCommitment is not deterministic")
	}
}

// TestInputSensitivity verifies that changing any single input changes the output.
func TestInputSensitivity(t *testing.T) {
	e := testElements(t)
	masterKey, invoiceId := e[0], e[1]
	altMasterKey := e[2]

	r1, err := DeriveNullifierKey(masterKey, invoiceId)
	if err != nil {
		t.Fatal(err)
	}

	r2, err := DeriveNullifierKey(altMasterKey, invoiceId)
	if err != nil {
		t.Fatal(err)
	}

	if r1.Equal(&r2) {
		t.Error("DeriveNullifierKey: different masterKey must produce different output")
	}

	altInvoiceId := e[3]
	r3, err := DeriveNullifierKey(masterKey, altInvoiceId)
	if err != nil {
		t.Fatal(err)
	}

	if r1.Equal(&r3) {
		t.Error("DeriveNullifierKey: different invoiceId must produce different output")
	}
}

// TestDeriveCreditNullifierKey verifies the formula matches a direct Poseidon2Hash call.
func TestDeriveCreditNullifierKey(t *testing.T) {
	e := testElements(t)
	receiverMasterKey, creditId := e[0], e[1]

	got, err := DeriveCreditNullifierKey(receiverMasterKey, creditId)
	if err != nil {
		t.Fatal(err)
	}

	want, err := Poseidon2Hash(ElementFromUint64(DomainCreditNKey), receiverMasterKey, creditId)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(&want) {
		t.Errorf("DeriveCreditNullifierKey: got %s, want %s", ElementToHex(got), ElementToHex(want))
	}
}

// TestDeriveCreditNullifierKeyDomainIsolation verifies that creditNullifierKey and
// the receiver/payer nullifierKey — same inputs, different domain tags (10 vs 1) —
// are distinct: holding one must not reveal the other.
func TestDeriveCreditNullifierKeyDomainIsolation(t *testing.T) {
	e := testElements(t)
	masterKey, creditId := e[0], e[1]

	creditKey, err := DeriveCreditNullifierKey(masterKey, creditId)
	if err != nil {
		t.Fatal(err)
	}
	nullifierKey, err := DeriveNullifierKey(masterKey, creditId)
	if err != nil {
		t.Fatal(err)
	}

	if creditKey.Equal(&nullifierKey) {
		t.Error("domain isolation failure: DeriveCreditNullifierKey and DeriveNullifierKey returned same value")
	}
}
