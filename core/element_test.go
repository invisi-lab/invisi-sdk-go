package core

import (
	"math/big"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestHexToElement(t *testing.T) {
	e, err := HexToElement("0x000000000000000000000000000000000000000000000000000000000000002a")
	if err != nil {
		t.Fatal(err)
	}
	var expected fr.Element
	expected.SetUint64(42)
	if !e.Equal(&expected) {
		t.Errorf("got %s, want 42", e.String())
	}
}

func TestHexToElementShort(t *testing.T) {
	e, err := HexToElement("0x2a")
	if err != nil {
		t.Fatal(err)
	}
	var expected fr.Element
	expected.SetUint64(42)
	if !e.Equal(&expected) {
		t.Errorf("got %s, want 42", e.String())
	}
}

func TestElementToHex(t *testing.T) {
	var e fr.Element
	e.SetUint64(42)
	hex := ElementToHex(e)
	if len(hex) != 66 {
		t.Errorf("hex length = %d, want 66", len(hex))
	}
	if hex[:2] != "0x" {
		t.Errorf("missing 0x prefix: %s", hex)
	}
	back, err := HexToElement(hex)
	if err != nil {
		t.Fatal(err)
	}
	if !back.Equal(&e) {
		t.Error("roundtrip failed")
	}
}

func TestBigIntToElement(t *testing.T) {
	b := big.NewInt(12345)
	e := BigIntToElement(b)
	back := ElementToBigInt(e)
	if back.Cmp(b) != 0 {
		t.Errorf("got %s, want %s", back, b)
	}
}

func TestElementFromUint64(t *testing.T) {
	e := ElementFromUint64(7)
	var expected fr.Element
	expected.SetUint64(7)
	if !e.Equal(&expected) {
		t.Errorf("ElementFromUint64(7) mismatch")
	}
}

func TestHexToElementInvalid(t *testing.T) {
	_, err := HexToElement("not-hex")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestHexToElementTooLong(t *testing.T) {
	// 40 bytes: previously drove copy(padded[32-len(b):], b) to a negative
	// offset and panicked; must now return an error instead.
	_, err := HexToElement("0x" + strings.Repeat("ff", 40))
	if err == nil {
		t.Error("expected error for hex longer than 32 bytes")
	}
}

func TestHexToElementBoundary(t *testing.T) {
	// 32 bytes is the largest accepted width; 33 bytes is the first rejected one.
	if _, err := HexToElement("0x" + strings.Repeat("00", 32)); err != nil {
		t.Errorf("32-byte hex should decode, got error: %v", err)
	}
	if _, err := HexToElement("0x" + strings.Repeat("ff", 33)); err == nil {
		t.Error("expected error for 33-byte hex")
	}
}

func TestHexToElementUppercasePrefix(t *testing.T) {
	e, err := HexToElement("0X2A")
	if err != nil {
		t.Fatal(err)
	}
	var expected fr.Element
	expected.SetUint64(42)
	if !e.Equal(&expected) {
		t.Errorf("got %s, want 42", e.String())
	}
}
