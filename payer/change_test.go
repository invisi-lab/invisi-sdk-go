package payer

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func changeFixture(t *testing.T) DeriveChangeParams {
	t.Helper()
	var payerMasterKey, nonce fr.Element
	payerMasterKey.SetUint64(12345)
	nonce.SetUint64(555)
	return DeriveChangeParams{
		PayerMasterKey: payerMasterKey,
		// PaymentCommitment must be a canonical BN254 field value (a real PC is a
		// Poseidon2 output). core.HexToElement uses SetBytesCanonical and rejects
		// out-of-field bytes, so DO NOT use 0xabab… (first byte 0xab > modulus).
		PaymentCommitment:      common.BigToHash(big.NewInt(67890)),
		ChangeReceivingAddress: common.HexToAddress("0x" + strings.Repeat("de", 20)),
		TokenAddress:           common.HexToAddress("0x" + strings.Repeat("aa", 20)),
		ChangeAmount:           big.NewInt(53_000_000),
		Nonce:                  nonce,
		ChainID:                big.NewInt(84532),
	}
}

func TestDeriveChange_FormulaLock(t *testing.T) {
	p := changeFixture(t)
	got, err := DeriveChange(p)
	if err != nil {
		t.Fatal(err)
	}

	pcElem, err := core.HexToElement(p.PaymentCommitment.Hex())
	if err != nil {
		t.Fatal(err)
	}
	wantKey, err := core.DeriveChangeNullifierKey(p.PayerMasterKey, pcElem)
	if err != nil {
		t.Fatal(err)
	}
	if !got.ChangeNullifierKey.Equal(&wantKey) {
		t.Errorf("changeNullifierKey mismatch")
	}

	addrElem := core.BigIntToElement(new(big.Int).SetBytes(p.ChangeReceivingAddress.Bytes()))
	tokenElem := core.BigIntToElement(new(big.Int).SetBytes(p.TokenAddress.Bytes()))
	wantCC, err := core.ComputeChangeCommitment(
		core.BigIntToElement(p.ChainID), wantKey, addrElem, tokenElem,
		core.BigIntToElement(p.ChangeAmount), p.Nonce, // changeNonce = payment nonce
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.ChangeCommitment != common.HexToHash(core.ElementToHex(wantCC)) {
		t.Errorf("changeCommitment mismatch: got %s", got.ChangeCommitment.Hex())
	}
}

func TestDeriveChange_Deterministic(t *testing.T) {
	p := changeFixture(t)
	a, err := DeriveChange(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DeriveChange(p)
	if err != nil {
		t.Fatal(err)
	}
	if !a.ChangeNullifierKey.Equal(&b.ChangeNullifierKey) || a.ChangeCommitment != b.ChangeCommitment {
		t.Error("DeriveChange not deterministic")
	}
}

func TestDeriveChange_RequiresPayerMasterKey(t *testing.T) {
	p := changeFixture(t)
	p.PayerMasterKey = fr.Element{} // zero = missing; no random fallback
	_, err := DeriveChange(p)
	if !errors.Is(err, ErrPayerMasterKeyRequired) {
		t.Fatalf("expected ErrPayerMasterKeyRequired, got %v", err)
	}
}
