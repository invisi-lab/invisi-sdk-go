package shared

import (
	"context"
	"errors"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
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

// Ephemeral inventory uses the same canonical synchronization and plaintext checks as durable sessions.
func discover(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, public *sdk.Wallet, decryptor sdk.DeliveryDecryptor) (walletsession.WalletState, error) {
	store := walletsession.NewMemoryWalletStore()
	managed, err := walletsession.Connect(ctx, walletsession.Options{Invisi: sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client}), Wallet: public, Protocol: pool, Store: store})
	if err != nil {
		return walletsession.WalletState{}, err
	}
	pending, err := managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock})
	if err != nil {
		return walletsession.WalletState{}, err
	}
	if _, err = pending.Decrypt(ctx, decryptor); err != nil {
		return walletsession.WalletState{}, err
	}

	return store.ReadSnapshot(ctx, managed.Scope())
}
func DiscoverNotes(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, public *sdk.Wallet, decryptor sdk.DeliveryDecryptor, token common.Address) ([]NoteCandidate, error) {
	state, err := discover(ctx, provider, client, pool, public, decryptor)
	if err != nil {
		return nil, err
	}
	result := []NoteCandidate{}
	for _, record := range state.Notes {
		if record.Spent || record.Note.TokenAddress != primitives.BigIntToElement(token.Big()) {
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
func DiscoverCredits(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, public *sdk.Wallet, decryptor sdk.DeliveryDecryptor, token common.Address, role string) ([]CreditCandidate, error) {
	selected := credit.Owner
	if role == "spender" {
		selected = credit.Spender
	} else if role != "owner" {
		return nil, errors.New("invalid Credit role")
	}
	state, err := discover(ctx, provider, client, pool, public, decryptor)
	if err != nil {
		return nil, err
	}
	return CreditCandidates(ctx, provider, client, pool, state, token, selected)
}

// CreditCandidates reuses a completed refresh so durable sessions do not ask the host to decrypt twice.
func CreditCandidates(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, state walletsession.WalletState, token common.Address, selected credit.Role) ([]CreditCandidate, error) {
	if state.DiscoveryThrough == nil {
		return nil, errors.New("canonical discovery anchor missing")
	}
	result := []CreditCandidate{}
	for _, record := range state.PostpaidCredit.Grants {
		if record.Role != selected || record.ReservedBy != "" || record.Consumed || record.Revoked || primitives.ElementToBigInt(record.Delivery.State.Expiry).Uint64() <= state.DiscoveryThrough.Timestamp || record.Delivery.State.TokenAddress != primitives.BigIntToElement(token.Big()) {
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
