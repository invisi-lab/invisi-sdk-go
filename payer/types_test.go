package payer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestGasEstimates(t *testing.T) {
	if GasEstimatePayOpen.CallGasLimit != 150_000 {
		t.Errorf("PayOpen.CallGasLimit = %d, want 150000", GasEstimatePayOpen.CallGasLimit)
	}
	if GasEstimatePayBlind.CallGasLimit != 350_000 {
		t.Errorf("PayBlind.CallGasLimit = %d, want 350000", GasEstimatePayBlind.CallGasLimit)
	}
	if GasEstimateDepositCredit.CallGasLimit != 180_000 {
		t.Errorf("DepositCredit.CallGasLimit = %d, want 180000", GasEstimateDepositCredit.CallGasLimit)
	}
}

func TestPayOpenParamsDefaults(t *testing.T) {
	// RefundLink and Nonce should be nil pointers (optional)
	p := PayOpenParams{
		ChainID:            big.NewInt(1),
		IdentityCommitment: common.HexToHash("0xab"),
		TokenAddress:       common.HexToAddress("0x1234"),
		Amount:             big.NewInt(100),
		TreeID:             big.NewInt(1),
	}
	if p.RefundLink != nil {
		t.Error("RefundLink should default to nil")
	}
	if p.Nonce != nil {
		t.Error("Nonce should default to nil")
	}
}

func TestCreditAuthorizationFields(t *testing.T) {
	// Verify CreditAuthorization holds all required fields
	auth := CreditAuthorization{}
	// Zero values should be valid (no panic)
	if auth.CreditID != (common.Hash{}) {
		t.Error("zero CreditID unexpected")
	}
	_ = auth.IdentityCommitment
	_ = auth.PayerCommitment
	_ = auth.CreditNullifierKey
	_ = auth.NullifierKeyHash
	_ = auth.TokenAddress
	_ = auth.Amount
	_ = auth.Nonce
	_ = auth.Expiry
	_ = auth.TreeID
	_ = auth.ChainID
}
