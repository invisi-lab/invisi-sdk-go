package deployments_test

import (
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/deployments"
)

func TestGetDeployment_ValidChainIDs(t *testing.T) {
	cases := []struct {
		chainID int64
		network string
		env     string
	}{
		{1, "mainnet", "mainnet"},
		{10, "optimism", "mainnet"},
		{8453, "base", "mainnet"},
		{42161, "arbitrum", "mainnet"},
		{11155111, "sepolia", "testnet"},
		{11155420, "optimism-sepolia", "testnet"},
		{84532, "base-sepolia", "testnet"},
		{421614, "arbitrum-sepolia", "testnet"},
	}

	for _, tc := range cases {
		t.Run(tc.network, func(t *testing.T) {
			d, err := deployments.GetDeployment(tc.chainID)
			if err != nil {
				t.Fatalf("GetDeployment(%d): unexpected error: %v", tc.chainID, err)
			}
			if d.ChainID != tc.chainID {
				t.Errorf("ChainID: got %d, want %d", d.ChainID, tc.chainID)
			}
			if d.Network != tc.network {
				t.Errorf("Network: got %q, want %q", d.Network, tc.network)
			}
			if d.Env != tc.env {
				t.Errorf("Env: got %q, want %q", d.Env, tc.env)
			}
		})
	}
}

func TestGetDeployment_InvalidChainID(t *testing.T) {
	invalidIDs := []int64{0, 2, 999, -1, 1337}
	for _, id := range invalidIDs {
		_, err := deployments.GetDeployment(id)
		if err == nil {
			t.Errorf("GetDeployment(%d): expected error, got nil", id)
		}
	}
}

func TestGetSupportedChainIDs(t *testing.T) {
	ids := deployments.GetSupportedChainIDs()

	if len(ids) != 8 {
		t.Errorf("GetSupportedChainIDs: expected 8 entries, got %d", len(ids))
	}

	// Verify sorted order.
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("GetSupportedChainIDs: not sorted at index %d: %d >= %d", i, ids[i-1], ids[i])
		}
	}

	// Verify first and last values (sorted ascending).
	if len(ids) > 0 && ids[0] != 1 {
		t.Errorf("GetSupportedChainIDs: expected first ID to be 1, got %d", ids[0])
	}
}

func TestGetSupportedChainIDs_AllValid(t *testing.T) {
	ids := deployments.GetSupportedChainIDs()
	for _, id := range ids {
		if _, err := deployments.GetDeployment(id); err != nil {
			t.Errorf("GetDeployment(%d) failed for a supported chain: %v", id, err)
		}
	}
}
