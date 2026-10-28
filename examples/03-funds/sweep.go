package funds

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

// Sweep prepares explicit inputs without a Store, Connect or Refresh.
func Sweep(ctx context.Context, provider sdk.Provider, protocolRPC *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.Wallet) (sdk.TransactionCall, error) {

	localProver, err := native.NewNativeProver(native.NativeOptions{})
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: protocolRPC, Prover: localProver})
	tokenAddress := common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6")

	// Reconstruct an EXISTING received note. Replace these sample fields and tree position
	// with your private record and its canonical insertion event; never replace its nonce.
	note := core.Note{ProtocolContext: core.ProtocolContext{
		ChainID:         core.ElementFromUint64(pool.Context().ChainID),
		ProtocolAddress: core.BigIntToElement(pool.Context().ProtocolAddress.Big()),
	}, WalletIdentity: core.BigIntToElement(wallet.WalletIdentity().Big()),
		TokenAddress: core.BigIntToElement(tokenAddress.Big()), Amount: core.ElementFromUint64(100), Nonce: core.ElementFromUint64(123)}
	identity, err := core.ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	leaf, err := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	inputs := []sdk.NoteInput{{Note: note, Commitment: common.Hash(leaf.Bytes()), TreeID: 1, LeafIndex: 0}}
	// Full payout to the bound EVM owner needs no private wallet signature.
	prepared, err := invisi.PrepareSweep(ctx, sdk.ExplicitSweepPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: wallet, Protocol: pool}, ExplicitFunds: sdk.ExplicitFunds{Inputs: inputs},
		SweepPreparation: sdk.SweepPreparation{TokenAddress: tokenAddress}})

	if err != nil {
		return sdk.TransactionCall{}, err
	}
	fmt.Println("Review the complete frozen operation locally:", prepared.Preview())
	reference, err := prepared.EstimateGas()
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	fmt.Println("Before proving: reference gas units", reference.GasUnits)
	authorized := prepared

	proved, err := authorized.Prove(ctx, prover.Options{OnProgress: func(event prover.Progress) { fmt.Println("Proof phase:", event.Phase) }})
	if err != nil {
		return sdk.TransactionCall{}, err
	}

	// Export only the public call. The application owns input reservations, sending,
	// canonical receipt checks and private inventory updates in this store-free mode.
	return proved.ExportTransaction(ctx)
}
