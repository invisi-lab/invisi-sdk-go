package payer

import (
	"math/big"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-labs/invisi-sdk-go/core"
)

func TestBuildSettleChangeCalldata_Basic(t *testing.T) {
	var nullifierKey fr.Element
	nullifierKey.SetUint64(42)
	var nonce fr.Element
	nonce.SetUint64(12345)

	result, err := BuildSettleChangeCalldata(SettleChangeParams{
		ChainID:          big.NewInt(1),
		NullifierKey:     nullifierKey,
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		TokenAddress:     common.HexToAddress("0xaa" + strings.Repeat("00", 19)),
		Amount:           big.NewInt(53_000_000),
		Nonce:            nonce,
		MerkleRoot:       common.HexToHash("0xbb" + strings.Repeat("00", 31)),
		Proof:            []byte{0xde, 0xad, 0xbe, 0xef},
		TreeID:           big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Nullifier must be non-zero
	if result.Nullifier == (common.Hash{}) {
		t.Error("nullifier should not be zero")
	}

	// Data must start with withdraw selector
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestBuildSettleChangeCalldata_NullifierConsistency(t *testing.T) {
	var nullifierKey fr.Element
	nullifierKey.SetUint64(42)
	var nonce fr.Element
	nonce.SetUint64(12345)
	chainID := big.NewInt(1)
	tokenAddr := common.HexToAddress("0xaa" + strings.Repeat("00", 19))
	amount := big.NewInt(53_000_000)

	result, err := BuildSettleChangeCalldata(SettleChangeParams{
		ChainID:          chainID,
		NullifierKey:     nullifierKey,
		ReceivingAddress: common.HexToAddress("0xdead"),
		TokenAddress:     tokenAddr,
		Amount:           amount,
		Nonce:            nonce,
		MerkleRoot:       common.HexToHash("0xbb"),
		Proof:            []byte{0x01},
		TreeID:           big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Recompute nullifier via core
	chainElem := core.BigIntToElement(chainID)
	tokenHash := common.BytesToHash(tokenAddr.Bytes())
	tokenElem, _ := core.HexToElement(tokenHash.Hex())
	amountElem := core.BigIntToElement(amount)

	expectedNull, err := core.ComputePaymentNullifier(chainElem, nullifierKey, tokenElem, amountElem, nonce)
	if err != nil {
		t.Fatal(err)
	}

	expectedHash := common.HexToHash(core.ElementToHex(expectedNull))
	if result.Nullifier != expectedHash {
		t.Errorf("nullifier mismatch: got %s, want %s", result.Nullifier.Hex(), expectedHash.Hex())
	}
}
