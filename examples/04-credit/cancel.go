package credit

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

// Cancel uses explicit private records; business persistence belongs to the application.
func Cancel(ctx context.Context, provider sdk.Provider, protocolRPC *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.Wallet, counterparty *sdk.Wallet) (sdk.TransactionCall, error) {

	localProver, err := native.NewNativeProver(native.NativeOptions{})
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: protocolRPC, Prover: localProver})
	tokenAddress := common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6")
	// Replace these illustrative fields with the actual delivered Credit state.
	// The original owner/spender receiving-key snapshots and shared key must be preserved.
	owner, spender := wallet, counterparty
	creditState := core.CreditDeliveryState{Note: core.CreditNote{
		ProtocolContext:     core.ProtocolContext{ChainID: core.ElementFromUint64(pool.Context().ChainID), ProtocolAddress: core.BigIntToElement(pool.Context().ProtocolAddress.Big())},
		OwnerWalletIdentity: core.BigIntToElement(owner.WalletIdentity().Big()), SpenderWalletIdentity: core.BigIntToElement(spender.WalletIdentity().Big()),
		OwnerPayout: core.BigIntToElement(owner.OwnerAddress().Big()), SpenderPayout: core.BigIntToElement(spender.OwnerAddress().Big()),
		TokenAddress: core.BigIntToElement(tokenAddress.Big()), Amount: core.ElementFromUint64(100), Nonce: core.ElementFromUint64(678),
		Expiry: core.ElementFromUint64(1900000000), CreditSharedKey: core.ElementFromUint64(999)},
		OwnerReceivingKeys:   core.ReceivingKeys{ViewingPublicKey: owner.ViewingPublicKey(), ClueKey: owner.ClueKey()},
		SpenderReceivingKeys: core.ReceivingKeys{ViewingPublicKey: spender.ViewingPublicKey(), ClueKey: spender.ClueKey()}}
	keyHash, err := core.ComputeCreditSharedKeyHash(creditState.Note.CreditSharedKey)
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	creditNote := creditState.Note
	creditLeaf, err := core.ComputeCreditCommitment(creditNote.ChainID, creditNote.SpenderWalletIdentity, creditNote.OwnerWalletIdentity, creditNote.SpenderPayout, creditNote.OwnerPayout, creditNote.TokenAddress, creditNote.Amount, creditNote.Nonce, creditNote.Expiry, keyHash)
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	commitment := common.Hash(creditLeaf.Bytes())
	creditInputs := []sdk.CreditInput{{State: creditState, Commitment: commitment, TreeID: 2, LeafIndex: 0, Role: "owner"}}
	prepared, err := invisi.PrepareCancelCredit(ctx, sdk.ExplicitCancelCreditPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: wallet, Protocol: pool}, CreditInputs: creditInputs,
		CancelCreditPreparation: sdk.CancelCreditPreparation{TokenAddress: tokenAddress, Credit: commitment}})

	if err != nil {
		return sdk.TransactionCall{}, err
	}
	fmt.Println("Review the complete frozen operation locally:", prepared.Preview())
	reference, err := prepared.EstimateGas()
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	fmt.Println("Before proving: reference gas units", reference.GasUnits)
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		return sdk.TransactionCall{}, err
	}

	proved, err := authorized.Prove(ctx, prover.Options{OnProgress: func(event prover.Progress) { fmt.Println("Proof phase:", event.Phase) }})
	if err != nil {
		return sdk.TransactionCall{}, err
	}

	// Export only the public call. The application owns input reservations, sending,
	// canonical receipt checks and private inventory updates in this store-free mode.
	return proved.ExportTransaction(ctx)
}
