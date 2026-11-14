package protocol

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestDirectFundingKeepsUnsignedNonceGuards(t *testing.T) {
	payer := common.HexToAddress("0x1")
	other := common.HexToAddress("0x2")
	for _, test := range []struct {
		name            string
		nonce, deadline *big.Int
		caller          common.Address
		valid           bool
	}{
		{"direct", big.NewInt(0), big.NewInt(0), payer, true},
		{"nonce", big.NewInt(1), big.NewInt(0), payer, false},
		{"deadline", big.NewInt(0), big.NewInt(1), payer, false},
		{"other caller", big.NewInt(0), big.NewInt(0), other, false},
		{"nonce overflow", new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(0), payer, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Skipping the unused digest must preserve nonce, deadline and payer guards for unsigned funding.
			funding := FundingParams{Payer: payer, Amount: big.NewInt(1), SigNonce: test.nonce, Deadline: test.deadline}
			err := ValidateFundingSubmission(funding, &test.caller, big.NewInt(1))
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v: %v", test.valid, err)
			}
		})
	}
}
