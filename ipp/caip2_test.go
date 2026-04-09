package ipp

import (
	"math/big"
	"testing"
)

func TestChainIDToCAIP2(t *testing.T) {
	tests := []struct {
		name    string
		chainID *big.Int
		want    string
	}{
		{"mainnet", big.NewInt(1), "eip155:1"},
		{"base sepolia", big.NewInt(84532), "eip155:84532"},
		{"large chain ID", new(big.Int).SetUint64(999999999), "eip155:999999999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChainIDToCAIP2(tt.chainID)
			if got != tt.want {
				t.Errorf("ChainIDToCAIP2(%v) = %q, want %q", tt.chainID, got, tt.want)
			}
		})
	}
}

func TestCAIP2ToChainID(t *testing.T) {
	tests := []struct {
		name    string
		network string
		want    int64
	}{
		{"mainnet", "eip155:1", 1},
		{"base sepolia", "eip155:84532", 84532},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CAIP2ToChainID(tt.network)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Int64() != tt.want {
				t.Errorf("CAIP2ToChainID(%q) = %d, want %d", tt.network, got.Int64(), tt.want)
			}
		})
	}
}

func TestCAIP2ToChainIDRoundtrip(t *testing.T) {
	original := big.NewInt(42161)
	caip2 := ChainIDToCAIP2(original)
	parsed, err := CAIP2ToChainID(caip2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if original.Cmp(parsed) != 0 {
		t.Errorf("roundtrip failed: %v != %v", original, parsed)
	}
}

func TestCAIP2ToChainIDErrors(t *testing.T) {
	tests := []struct {
		name    string
		network string
	}{
		{"bad prefix", "cosmos:1"},
		{"no prefix", "1"},
		{"empty", ""},
		{"empty chain ID", "eip155:"},
		{"non-numeric", "eip155:abc"},
		{"float", "eip155:1.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CAIP2ToChainID(tt.network)
			if err == nil {
				t.Errorf("CAIP2ToChainID(%q) expected error, got nil", tt.network)
			}
		})
	}
}
