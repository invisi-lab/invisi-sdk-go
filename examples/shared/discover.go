package shared

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
)

type NoteCandidate struct {
	Input  sdk.NoteInput
	Output rpc.ProtocolNoteOutput
}
type CreditCandidate struct {
	Input      credit.LocatedState
	Delivery   credit.DeliveryState
	Commitment common.Hash
	Membership rpc.Membership
}

// DiscoverNotes scans genuine indexed outputs in MaxUpdateBlocks-bounded pages, authenticates decryption,
// rejects conflicting records and filters spent nullifiers against the chain. Preparation
// independently validates membership and the latest spent state; callers own reservations.
func DiscoverNotes(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.PublicWallet, viewingKey []byte, token common.Address) ([]NoteCandidate, error) {
	viewer, err := sdk.NewViewingCapability(wallet.OwnerAddress(), viewingKey, wallet.ViewingPublicKey())
	if err != nil {
		return nil, err
	}
	client = client.WithProvider(provider)
	d := pool.Context()
	deployment := rpc.DeploymentContext{ChainID: d.ChainID, ProtocolAddress: d.ProtocolAddress}
	status, err := client.GetStatus(ctx, deployment)
	if err != nil {
		return nil, err
	}
	if status.IndexedHead == nil {
		return nil, fmt.Errorf("Protocol RPC has no indexed head")
	}
	contract, err := contracts.NewInvisiProtocolCaller(d.ProtocolAddress, provider)
	if err != nil {
		return nil, err
	}
	seen := map[common.Hash]string{}
	positions := map[string]string{}
	result := []NoteCandidate{}
	for from := pool.Deployment().Protocol.FromBlock; from <= status.IndexedHead.Number; {
		to := min(from+rpc.MaxUpdateBlocks-1, status.IndexedHead.Number)
		page, err := client.GetNoteUpdates(ctx, deployment, rpc.ProtocolUpdateRange{FromBlock: from, ToBlock: to})
		if err != nil {
			return nil, err
		}
		for _, output := range page.Notes {
			encoded, _ := json.Marshal(output)
			fingerprint := string(encoded)
			position := fmt.Sprintf("%d:%d", output.TreeID, output.LeafIndex)
			if prior, ok := seen[output.Commitment]; ok {
				if prior != fingerprint {
					return nil, fmt.Errorf("conflicting note record")
				}
				continue
			}
			if prior, ok := positions[position]; ok && prior != fingerprint {
				return nil, fmt.Errorf("conflicting note tree position")
			}
			seen[output.Commitment] = fingerprint
			positions[position] = fingerprint
			if output.TokenAddress != token {
				continue
			}
			note, err := viewer.DecryptNote(output.Ciphertext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(d.ChainID), ProtocolAddress: core.BigIntToElement(d.ProtocolAddress.Big())}, OutputCommitment: core.BigIntToElement(output.Commitment.Big())}, TokenAddress: core.BigIntToElement(token.Big()), OwnerAddress: core.BigIntToElement(wallet.OwnerAddress().Big())})
			if err != nil {
				continue
			}
			nullifier, err := wallet.ExportNullifier().ForNote(note)
			if err != nil {
				return nil, err
			}
			spent, err := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx}, nullifier)
			if err != nil {
				return nil, err
			}
			if spent {
				continue
			}
			result = append(result, NoteCandidate{sdk.NoteInput{Note: note, Commitment: output.Commitment, TreeID: output.TreeID, LeafIndex: output.LeafIndex}, output})
		}
		if to == status.IndexedHead.Number {
			break
		}
		from = to + 1
	}
	return result, nil
}

// DiscoverCredits composes the same canonical receipt authentication used by managed sessions.
func DiscoverCredits(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.PublicWallet, viewingKey []byte, token common.Address, role string) ([]CreditCandidate, error) {
	selected := credit.Owner
	if role == "spender" {
		selected = credit.Spender
	} else if role != "owner" {
		return nil, fmt.Errorf("invalid grant role")
	}
	viewer, err := sdk.NewViewingCapability(wallet.OwnerAddress(), viewingKey, wallet.ViewingPublicKey())
	if err != nil {
		return nil, err
	}
	store := walletsession.NewMemoryWalletStore()
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client})
	session, err := walletsession.Connect(ctx, walletsession.Options{Invisi: invisi, Wallet: wallet, Protocol: pool, Store: store, Viewing: viewer})
	if err != nil {
		return nil, err
	}
	if _, err = session.Refresh(ctx, walletsession.RefreshOptions{}); err != nil {
		return nil, err
	}
	snapshot, err := store.ReadSnapshot(ctx, session.Scope())
	if err != nil {
		return nil, err
	}
	if snapshot.DiscoveryThrough == nil {
		return nil, fmt.Errorf("grant discovery anchor missing")
	}
	result := []CreditCandidate{}
	for _, record := range snapshot.PostpaidCredit.Grants {
		if record.Role != selected || record.Consumed || record.Revoked || core.ElementToBigInt(record.Delivery.State.TokenAddress).Cmp(token.Big()) != 0 {
			continue
		}
		membership, err := client.WithProvider(provider).GetMembership(ctx, pool.Context(), rpc.MembershipInput{TreeID: record.TreeID, LeafIndex: uint32(record.LeafIndex), Commitment: record.Commitment}, *snapshot.DiscoveryThrough)
		if err != nil {
			return nil, err
		}
		var path [16]common.Hash
		if len(membership.Path) != 16 {
			return nil, fmt.Errorf("grant path requires16siblings")
		}
		copy(path[:], membership.Path)
		result = append(result, CreditCandidate{Input: credit.LocatedState{State: record.Delivery.State, TreeID: record.TreeID, LeafIndex: record.LeafIndex, Path: path}, Delivery: record.Delivery, Commitment: record.Commitment, Membership: membership})
	}
	return result, nil
}
