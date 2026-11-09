package shared

import (
	"context"
	"errors"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
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

// Discovery uses canonical managed synchronization in an ephemeral inventory. The host
// explicitly executes each captured request batch; no viewing key enters this process.
func discover(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, public *sdk.Wallet, nativeWallet wallet.Client) (walletsession.WalletState, error) {
	store := walletsession.NewMemoryWalletStore()
	managed, err := walletsession.Connect(ctx, walletsession.Options{Invisi: sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client}), Wallet: public, Protocol: pool, Store: store})
	if err != nil {
		return walletsession.WalletState{}, err
	}
	pending, err := managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock})
	if err != nil {
		return walletsession.WalletState{}, err
	}
	responses, err := nativeWallet.Execute(ctx, pending.WalletRequests)
	if err != nil {
		return walletsession.WalletState{}, err
	}
	if _, err = pending.ApplyWalletResponses(ctx, responses); err != nil {
		return walletsession.WalletState{}, err
	}
	return store.ReadSnapshot(ctx, managed.Scope())
}
func DiscoverNotes(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, public *sdk.Wallet, nativeWallet wallet.Client, token common.Address) ([]NoteCandidate, error) {
	state, err := discover(ctx, provider, client, pool, public, nativeWallet)
	if err != nil {
		return nil, err
	}
	result := []NoteCandidate{}
	for _, record := range state.Notes {
		if record.Spent || record.Note.TokenAddress != core.BigIntToElement(token.Big()) {
			continue
		}
		result = append(result, NoteCandidate{Input: sdk.NoteInput{Note: record.Note, Commitment: record.Commitment, TreeID: record.TreeID, LeafIndex: record.LeafIndex}, Output: rpc.ProtocolNoteOutput{Commitment: record.Commitment, TokenAddress: token, TreeID: record.TreeID, LeafIndex: record.LeafIndex, TransactionHash: record.TransactionHash}})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Input.TreeID != result[j].Input.TreeID {
			return result[i].Input.TreeID < result[j].Input.TreeID
		}
		return result[i].Input.LeafIndex < result[j].Input.LeafIndex
	})
	return result, nil
}
func DiscoverCredits(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, public *sdk.Wallet, nativeWallet wallet.Client, token common.Address, role string) ([]CreditCandidate, error) {
	selected := credit.Owner
	if role == "spender" {
		selected = credit.Spender
	} else if role != "owner" {
		return nil, errors.New("invalid Credit role")
	}
	state, err := discover(ctx, provider, client, pool, public, nativeWallet)
	if err != nil {
		return nil, err
	}
	if state.DiscoveryThrough == nil {
		return nil, errors.New("canonical discovery anchor missing")
	}
	result := []CreditCandidate{}
	for _, record := range state.PostpaidCredit.Grants {
		if record.Role != selected || record.Consumed || record.Revoked || core.ElementToBigInt(record.Delivery.State.Expiry).Uint64() <= state.DiscoveryThrough.Timestamp || record.Delivery.State.TokenAddress != core.BigIntToElement(token.Big()) {
			continue
		}
		membership, err := client.WithProvider(provider).GetMembership(ctx, pool.Context(), rpc.MembershipInput{TreeID: record.TreeID, LeafIndex: uint32(record.LeafIndex), Commitment: record.Commitment}, *state.DiscoveryThrough)
		if err != nil {
			return nil, err
		}
		if len(membership.Path) != 16 {
			return nil, errors.New("Credit membership depth mismatch")
		}
		var path [16]common.Hash
		copy(path[:], membership.Path)
		result = append(result, CreditCandidate{Input: credit.LocatedState{State: record.Delivery.State, TreeID: record.TreeID, LeafIndex: record.LeafIndex, Path: path}, Delivery: record.Delivery, Commitment: record.Commitment, Membership: membership})
	}
	return result, nil
}
