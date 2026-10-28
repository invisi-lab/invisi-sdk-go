package aggregation

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
)

// Aggregate supplies four disjoint notes from the application's own inventory.
// No Store or WalletSession is required. Keep application reservations across uncertain outcomes.
func Aggregate(ctx context.Context, provider sdk.Provider, protocolRPC *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.Wallet, recipient string, backend sdk.Prover) (sdk.TransactionCall, error) {
	aggregator, err := prover.NewAggregator(backend)
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: protocolRPC, Prover: backend})
	tokenAddress := common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6")
	members := []prover.AggregateMember{}
	requests := []protocol.TransactRequest{}
	fmt.Println("1. Independently authorize and prove four recursive slots14 members")
	for i := range 4 {
		// Illustrative EXISTING notes only. Replace every preimage and canonical position
		// with actual received records. The SDK cannot turn fabricated notes into funds.
		note := core.Note{ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(pool.Context().ChainID), ProtocolAddress: core.BigIntToElement(pool.Context().ProtocolAddress.Big())},
			WalletIdentity: core.BigIntToElement(wallet.WalletIdentity().Big()), TokenAddress: core.BigIntToElement(tokenAddress.Big()), Amount: core.ElementFromUint64(100), Nonce: core.ElementFromUint64(uint64(123 + i))}
		identity, err := core.ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
		if err != nil {
			return sdk.TransactionCall{}, err
		}
		leaf, err := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
		if err != nil {
			return sdk.TransactionCall{}, err
		}
		prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
			TransactionIdentity: sdk.TransactionIdentity{Wallet: wallet, Protocol: pool},
			ExplicitFunds:       sdk.ExplicitFunds{Inputs: []sdk.NoteInput{{Note: note, Commitment: common.Hash(leaf.Bytes()), TreeID: 1, LeafIndex: uint32(i)}}},
			TransferPreparation: sdk.TransferPreparation{TokenAddress: tokenAddress, Outputs: []sdk.PaymentOutput{{To: recipient, Amount: big.NewInt(1)}}, Target: "noir-recursive"}})
		if err != nil {
			return sdk.TransactionCall{}, err
		}
		fmt.Println("Review member:", prepared.Preview())
		authorized, err := prepared.Authorize(ctx)
		if err != nil {
			return sdk.TransactionCall{}, err
		}
		ready, err := authorized.Prove(ctx, prover.Options{OnProgress: func(event prover.Progress) { fmt.Println("Proof phase:", event.Phase) }})
		if err != nil {
			return sdk.TransactionCall{}, err
		}
		member, err := ready.ExportAggregateMember(ctx)
		if err != nil {
			return sdk.TransactionCall{}, err
		}
		members = append(members, prover.AggregateMember{Bundle: member.Bundle, Statement: member.Statement})
		requests = append(requests, member.Request)
	}
	fmt.Println("2. Aggregate public member proofs and statements; private witness data stays local")
	proof, err := aggregator.Aggregate(ctx, members, prover.Options{OnProgress: func(event prover.Progress) { fmt.Println("Aggregate proof phase:", event.Phase) }})
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	fmt.Println("3. Export the atomic public call for the application's ordinary geth pipeline")
	return pool.Aggregate(protocol.AggregateRequest{Proof: proof.Proof, Requests: requests})
}
