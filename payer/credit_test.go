package payer

import (
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestCreateCredit_Basic(t *testing.T) {
	var masterKey, creditID, payerAddr, nonce fr.Element
	masterKey.SetUint64(111)
	creditID.SetUint64(222)
	payerAddr.SetUint64(333)
	nonce.SetUint64(555)

	result, err := CreateCredit(CreateCreditParams{
		PayerMasterKey:     masterKey,
		CreditID:           creditID,
		IdentityCommitment: common.HexToHash("0xab" + strings.Repeat("00", 31)),
		PayerAddress:       payerAddr,
		TokenAddress:       common.HexToAddress("0xcc" + strings.Repeat("00", 19)),
		Amount:             big.NewInt(1_000_000),
		Nonce:              nonce,
		Expiry:             big.NewInt(1700000000),
		TreeID:             big.NewInt(1),
		ChainID:            big.NewInt(84532),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Authorization fields are populated
	if result.Authorization.PayerCommitment == (common.Hash{}) {
		t.Error("payerCommitment should not be zero")
	}
	if result.Authorization.NullifierKeyHash == (common.Hash{}) {
		t.Error("nullifierKeyHash should not be zero")
	}

	// Nonce must equal the supplied (authorize-response) value
	if !result.Authorization.Nonce.Equal(&nonce) {
		t.Errorf("nonce = %s, want supplied value", result.Authorization.Nonce.String())
	}

	// ApproveAmount == Amount
	if result.ApproveAmount.Cmp(big.NewInt(1_000_000)) != 0 {
		t.Errorf("approveAmount = %s, want 1000000", result.ApproveAmount)
	}

	// Data is valid calldata
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestCreateCredit_PayerCommitmentConsistency(t *testing.T) {
	var masterKey, creditID, payerAddr, nonce fr.Element
	masterKey.SetUint64(111)
	creditID.SetUint64(222)
	payerAddr.SetUint64(333)
	nonce.SetUint64(1)

	result, err := CreateCredit(CreateCreditParams{
		PayerMasterKey:     masterKey,
		CreditID:           creditID,
		IdentityCommitment: common.HexToHash("0xab" + strings.Repeat("00", 31)),
		PayerAddress:       payerAddr,
		TokenAddress:       common.HexToAddress("0xaa"),
		Amount:             big.NewInt(100),
		Nonce:              nonce,
		Expiry:             big.NewInt(999999),
		TreeID:             big.NewInt(1),
		ChainID:            big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Recompute payerCommitment = ComputeIdentityCommitment(payerNullifierKey, payerAddress)
	payerNullifierKey, err := core.DeriveNullifierKey(masterKey, creditID)
	if err != nil {
		t.Fatal(err)
	}
	expectedPayerComm, err := core.ComputeIdentityCommitment(payerNullifierKey, payerAddr)
	if err != nil {
		t.Fatal(err)
	}

	expectedHash := common.HexToHash(core.ElementToHex(expectedPayerComm))
	if result.Authorization.PayerCommitment != expectedHash {
		t.Errorf("payerCommitment mismatch: got %s, want %s",
			result.Authorization.PayerCommitment.Hex(), expectedHash.Hex())
	}
}

func TestCreateCredit_ChainIDPreserved(t *testing.T) {
	var masterKey, creditID, payerAddr, nonce fr.Element
	masterKey.SetUint64(1)
	creditID.SetUint64(2)
	payerAddr.SetUint64(3)
	nonce.SetUint64(1)

	result, err := CreateCredit(CreateCreditParams{
		PayerMasterKey:     masterKey,
		CreditID:           creditID,
		IdentityCommitment: common.HexToHash("0x01"),
		PayerAddress:       payerAddr,
		TokenAddress:       common.HexToAddress("0x01"),
		Amount:             big.NewInt(100),
		Nonce:              nonce,
		Expiry:             big.NewInt(999),
		TreeID:             big.NewInt(1),
		ChainID:            big.NewInt(42161),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Authorization.ChainID.Int64() != 42161 {
		t.Errorf("chainId = %d, want 42161", result.Authorization.ChainID.Int64())
	}
}

func TestDeriveCreditAuthorizeFields(t *testing.T) {
	// Reuse the same fixture inputs as TestCreateCredit_Basic so the helper's
	// derivation chain can be checked byte-for-byte against CreateCredit's output.
	var masterKey, creditID, payerAddr, nonce fr.Element
	masterKey.SetUint64(111)
	creditID.SetUint64(222)
	payerAddr.SetUint64(333)
	nonce.SetUint64(555)

	payerCommitment, nullifierKeyHash, err := DeriveCreditAuthorizeFields(masterKey, creditID, payerAddr)
	if err != nil {
		t.Fatal(err)
	}

	// Both outputs must match the bytes32 hex shape.
	hexRe := regexp.MustCompile(`^0x[0-9a-f]{64}$`)
	if !hexRe.MatchString(payerCommitment) {
		t.Errorf("payerCommitment %q does not match bytes32 hex shape", payerCommitment)
	}
	if !hexRe.MatchString(nullifierKeyHash) {
		t.Errorf("nullifierKeyHash %q does not match bytes32 hex shape", nullifierKeyHash)
	}

	// Helper must produce the identical values that CreateCredit derives internally,
	// proving they share the same derivation chain (incl. the D_DERIVE middle step).
	result, err := CreateCredit(CreateCreditParams{
		PayerMasterKey:     masterKey,
		CreditID:           creditID,
		IdentityCommitment: common.HexToHash("0xab" + strings.Repeat("00", 31)),
		PayerAddress:       payerAddr,
		TokenAddress:       common.HexToAddress("0xcc" + strings.Repeat("00", 19)),
		Amount:             big.NewInt(1_000_000),
		Nonce:              nonce,
		Expiry:             big.NewInt(1700000000),
		TreeID:             big.NewInt(1),
		ChainID:            big.NewInt(84532),
	})
	if err != nil {
		t.Fatal(err)
	}

	if payerCommitment != result.Authorization.PayerCommitment.Hex() {
		t.Errorf("payerCommitment mismatch: helper %s, CreateCredit %s",
			payerCommitment, result.Authorization.PayerCommitment.Hex())
	}
	if nullifierKeyHash != result.Authorization.NullifierKeyHash.Hex() {
		t.Errorf("nullifierKeyHash mismatch: helper %s, CreateCredit %s",
			nullifierKeyHash, result.Authorization.NullifierKeyHash.Hex())
	}
}

func TestBuildReclaimCalldata_Basic(t *testing.T) {
	result, err := BuildReclaimCalldata(ReclaimParams{
		Nullifier:        common.HexToHash("0xdd" + strings.Repeat("00", 31)),
		Amount:           big.NewInt(1_000_000),
		ReceivingAddress: common.HexToAddress("0xdead"),
		MerkleRoot:       common.HexToHash("0xee" + strings.Repeat("00", 31)),
		Proof:            []byte{0xde, 0xad},
		TreeID:           big.NewInt(1),
		Expiry:           big.NewInt(1700000000),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
	if result.Nullifier != common.HexToHash("0xdd"+strings.Repeat("00", 31)) {
		t.Error("nullifier should be passed through unchanged")
	}
}

func TestBuildReclaimCalldata_RequiresFields(t *testing.T) {
	_, err := BuildReclaimCalldata(ReclaimParams{
		Nullifier: common.HexToHash("0x01"),
		// Missing Amount, TreeID, Expiry
	})
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}
