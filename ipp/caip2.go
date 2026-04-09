package ipp

import (
	"fmt"
	"math/big"
	"strings"
)

const caip2Prefix = "eip155:"

// ChainIDToCAIP2 converts a numeric EVM chain ID to its CAIP-2 representation
// (e.g. "eip155:1").
func ChainIDToCAIP2(chainID *big.Int) string {
	return caip2Prefix + chainID.String()
}

// CAIP2ToChainID parses a CAIP-2 network string (e.g. "eip155:1") and returns
// the numeric chain ID.
func CAIP2ToChainID(network string) (*big.Int, error) {
	if !strings.HasPrefix(network, caip2Prefix) {
		return nil, fmt.Errorf("ipp: invalid CAIP-2 prefix: %q", network)
	}
	numStr := strings.TrimPrefix(network, caip2Prefix)
	if numStr == "" {
		return nil, fmt.Errorf("ipp: empty chain ID in CAIP-2 string: %q", network)
	}
	id := new(big.Int)
	if _, ok := id.SetString(numStr, 10); !ok {
		return nil, fmt.Errorf("ipp: non-numeric chain ID in CAIP-2 string: %q", network)
	}
	return id, nil
}
