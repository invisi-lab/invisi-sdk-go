package deployments

import (
	"fmt"
	"sort"
)

// registry maps chain ID to its deployment metadata.
// Addresses are zero-valued (placeholder) until the protocol is deployed on each chain.
var registry = map[int64]ChainDeployment{
	1:        {ChainID: 1, Network: "mainnet", Env: "mainnet"},
	10:       {ChainID: 10, Network: "optimism", Env: "mainnet"},
	8453:     {ChainID: 8453, Network: "base", Env: "mainnet"},
	42161:    {ChainID: 42161, Network: "arbitrum", Env: "mainnet"},
	11155111: {ChainID: 11155111, Network: "sepolia", Env: "testnet"},
	11155420: {ChainID: 11155420, Network: "optimism-sepolia", Env: "testnet"},
	84532:    {ChainID: 84532, Network: "base-sepolia", Env: "testnet"},
	421614:   {ChainID: 421614, Network: "arbitrum-sepolia", Env: "testnet"},
}

// GetDeployment returns the ChainDeployment for the given chain ID.
// Returns an error if the chain ID is not supported.
func GetDeployment(chainID int64) (ChainDeployment, error) {
	d, ok := registry[chainID]
	if !ok {
		return ChainDeployment{}, fmt.Errorf("unsupported chain ID: %d", chainID)
	}
	// Return independent mutable collections, keeping the registry immutable.
	d.TransactUserVerifiers = make(map[int]ContractDeployment)
	d.TransactAggregateVerifiers = make(map[int]ContractDeployment)
	d.CreditVerifiers = make(map[string]ContractDeployment)
	d.CreationCode = append([]byte(nil), d.CreationCode...)
	return d, nil
}

// GetSupportedChainIDs returns all registered chain IDs in ascending order.
func GetSupportedChainIDs() []int64 {
	ids := make([]int64, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
