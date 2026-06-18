package receiver

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestGenerateInvoice_ReturnsNullifierKeyAndIdentityCommitment(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	inv, err := GenerateInvoice(GenerateInvoiceParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		TokenAddress:     common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Amount:           big.NewInt(100_000_000),
		TreeID:           big.NewInt(1),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// NullifierKey must not be zero
	var zero fr.Element
	if inv.NullifierKey.Equal(&zero) {
		t.Error("NullifierKey is zero")
	}

	// IdentityCommitment must not be zero hash
	if inv.IdentityCommitment == (common.Hash{}) {
		t.Error("IdentityCommitment is zero hash")
	}

	// Verify against core functions
	expectedNK, err := core.DeriveNullifierKey(mk, invID)
	if err != nil {
		t.Fatal(err)
	}
	if !inv.NullifierKey.Equal(&expectedNK) {
		t.Errorf("NullifierKey mismatch: got %s, want %s",
			core.ElementToHex(inv.NullifierKey), core.ElementToHex(expectedNK))
	}

	recvAddrBigInt := new(big.Int).SetBytes(common.HexToAddress("0x000000000000000000000000000000000000dead").Bytes())
	recvAddrElem := core.BigIntToElement(recvAddrBigInt)
	expectedIC, err := core.ComputeIdentityCommitment(expectedNK, recvAddrElem)
	if err != nil {
		t.Fatal(err)
	}
	expectedICHash := common.HexToHash(core.ElementToHex(expectedIC))
	if inv.IdentityCommitment != expectedICHash {
		t.Errorf("IdentityCommitment mismatch: got %s, want %s",
			inv.IdentityCommitment.Hex(), expectedICHash.Hex())
	}
}

func TestGenerateInvoice_DifferentAddressesProduceDifferentCommitments(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(1)
	var invID fr.Element
	invID.SetUint64(1)

	a, err := GenerateInvoice(GenerateInvoiceParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		TokenAddress:     common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Amount:           big.NewInt(100_000_000),
		TreeID:           big.NewInt(1),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
	})
	if err != nil {
		t.Fatal(err)
	}

	b, err := GenerateInvoice(GenerateInvoiceParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000beef"),
		TokenAddress:     common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Amount:           big.NewInt(100_000_000),
		TreeID:           big.NewInt(1),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Same masterKey + invoiceId → same nullifierKey
	if !a.NullifierKey.Equal(&b.NullifierKey) {
		t.Error("NullifierKey should be the same for same masterKey+invoiceId")
	}

	// Different receivingAddress → different identityCommitment
	if a.IdentityCommitment == b.IdentityCommitment {
		t.Error("IdentityCommitment should differ for different receivingAddress")
	}
}

func TestGenerateInvoice_PassthroughFields(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(99)

	tokenAddr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	opAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recvAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")

	inv, err := GenerateInvoice(GenerateInvoiceParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: recvAddr,
		TokenAddress:     tokenAddr,
		Amount:           big.NewInt(100_000_000),
		TreeID:           big.NewInt(3),
		Operator:         opAddr,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !inv.InvoiceID.Equal(&invID) {
		t.Error("InvoiceID mismatch")
	}
	if inv.ReceivingAddress != recvAddr {
		t.Error("ReceivingAddress mismatch")
	}
	if inv.TokenAddress != tokenAddr {
		t.Error("TokenAddress mismatch")
	}
	if inv.Amount.Cmp(big.NewInt(100_000_000)) != 0 {
		t.Errorf("Amount = %s, want 100000000", inv.Amount)
	}
	if inv.TreeID.Cmp(big.NewInt(3)) != 0 {
		t.Errorf("TreeID = %s, want 3", inv.TreeID)
	}
	if inv.Operator != opAddr {
		t.Error("Operator mismatch")
	}
}
