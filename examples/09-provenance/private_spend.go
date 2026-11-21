// Package provenanceexample demonstrates Source recognition and private Note reuse.
package provenanceexample

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/examples/shared"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/provenance"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
)

// Recognition preserves the service's real states/reasons, including pending.
// A successful Router receipt does not imply recognized provenance.
func Recognition(ctx context.Context, source *provenance.Client, tx, commitment common.Hash) (provenance.Result, provenance.Result, error) {
	if source == nil {
		return nil, nil, fmt.Errorf("trusted Source client required")
	}
	transaction, err := source.GetTransactionRecognition(ctx, tx.Hex())
	if err != nil {
		return nil, nil, err
	}
	note, err := source.GetNoteRecognition(ctx, commitment.Hex())
	return transaction, note, err
}

// Spend selects a real unspent recognized Note, transfers half, and returns change.
// invisi must have real Core and Provenance provers plus the same Source capability.
// The decryptor and wallet signer are App capabilities, never fabricated examples.
// Persistence must atomically retain each callback's public artifacts before returning.
func Spend(ctx context.Context, invisi *sdk.Invisi, provider sdk.Provider,
	confirmation protocolrpc.TransactionConfirmationProvider, rpc *sdk.ProtocolRpcClient,
	pool *sdk.Protocol, wallet *sdk.Wallet, source *provenance.Client,
	decryptor sdk.DeliveryDecryptor, token common.Address, recipient string, deadline uint64,
	sign func(context.Context, primitives.SigningRequest) ([]byte, error),
	persist func(protocol.TransactionCall, provenance.ProvenanceProof, provenance.Tracking, common.Hash) error,
	send func(context.Context, protocol.TransactionCall) (common.Hash, error), submit bool,
) (provenance.Tracking, error) {
	var tracking provenance.Tracking
	if invisi == nil || provider == nil || confirmation == nil || rpc == nil || pool == nil || wallet == nil || source == nil || decryptor == nil || sign == nil || persist == nil || (submit && send == nil) {
		return tracking, fmt.Errorf("real SDK, chain, Source, decryptor, wallet and persistence capabilities required")
	}
	candidates, err := shared.DiscoverNotes(ctx, provider, rpc, pool, wallet, decryptor, token)
	if err != nil {
		return tracking, err
	}
	var selected *sdk.NoteInput
	for _, candidate := range candidates {
		amount := primitives.ElementToBigInt(candidate.Input.Note.Amount)
		if amount.Cmp(big.NewInt(2)) < 0 {
			continue
		}
		// Discovery authenticates canonical receipt ciphertext and membership;
		// independently require finalized execution before selecting the Note.
		finality, err := shared.WaitNativeFinality(ctx, confirmation, candidate.Output.TransactionHash, 30*time.Second)
		if err != nil {
			return tracking, err
		}
		// Recheck the discovered position/commitment against the finalized canonical tree.
		if _, err = rpc.WithProvider(provider).GetMembership(ctx, pool.Context(), protocolrpc.MembershipInput{TreeID: candidate.Input.TreeID, LeafIndex: uint32(candidate.Input.LeafIndex), Commitment: candidate.Input.Commitment}, *finality.Block); err != nil {
			return tracking, err
		}
		recognition, err := source.GetNoteRecognition(ctx, candidate.Input.Commitment.Hex())
		if err != nil {
			return tracking, err
		}
		var found bool
		var state string
		if err = json.Unmarshal(recognition["found"], &found); err != nil {
			return tracking, err
		}
		// A missing Note need not carry a recognitionState.
		if !found {
			continue
		}
		if err = json.Unmarshal(recognition["recognitionState"], &state); err != nil {
			return tracking, err
		}
		if state == "recognized" {
			input := candidate.Input
			selected = &input
			break
		}
	}
	if selected == nil {
		return tracking, fmt.Errorf("no finalized recognized unspent Note with at least two units")
	}
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
		ChainID: new(big.Int).SetUint64(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress,
		TokenAddress: token, Amount: new(big.Int).Div(primitives.ElementToBigInt(selected.Note.Amount), big.NewInt(2)),
		Recipient: sdk.InvisiRecipient{Address: recipient},
	})
	if err != nil {
		return tracking, err
	}
	changeAddress, err := wallet.GetAddress()
	if err != nil {
		return tracking, err
	}
	prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: wallet, Protocol: pool},
		ExplicitFunds:       sdk.ExplicitFunds{Inputs: []sdk.NoteInput{*selected}, ChangeAddress: changeAddress},
		TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}, Deadline: deadline},
	})
	if err != nil {
		return tracking, err
	}
	request, err := prepared.ExportWalletAuthorization(ctx)
	if err != nil {
		return tracking, err
	}
	signature, err := sign(ctx, request)
	if err != nil {
		return tracking, err
	}
	prepared, err = prepared.WithWalletSignature(ctx, signature)
	if err != nil {
		return tracking, err
	}
	prepared, err = prepared.Prove(ctx) // Core proof: main transaction validity.
	if err != nil {
		return tracking, err
	}
	prepared, err = prepared.ProveProvenance(ctx) // Source proof: existing 11 public inputs.
	if err != nil {
		return tracking, err
	}
	packet, err := prepared.ExportProvenanceProof(ctx)
	if err != nil {
		return tracking, err
	}
	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		return tracking, err
	}
	if !bytes.Equal(call.Data, common.FromHex(packet.Data)) {
		return tracking, fmt.Errorf("Core calldata differs from Source packet")
	}
	preflight, err := source.EvaluateTransaction(ctx, packet.Data, packet.Sidecars)
	if err != nil {
		return tracking, err
	}
	// One direct Core transact has SDK-assigned operation index 0.
	// Never infer a nested Router index from output or array positions.
	tracking, err = provenance.PreflightTracking(preflight, "0", packet.OperationDigest)
	if err != nil {
		return tracking, err
	}
	if err = persist(call, packet, tracking, common.Hash{}); err != nil {
		return tracking, err
	}
	if !submit {
		return tracking, nil
	}
	accepted, err := sdk.SubmitProvenanceProof(ctx, source, packet)
	if err != nil {
		return tracking, err
	}
	tracking, err = provenance.PreflightTracking(accepted, "0", packet.OperationDigest)
	if err != nil {
		return tracking, err
	}
	if err = persist(call, packet, tracking, common.Hash{}); err != nil {
		return tracking, err
	}
	hash, err := send(ctx, call) // Executor pays gas; principal remains the old Note.
	if err != nil {
		return tracking, err
	}
	if err = persist(call, packet, tracking, hash); err != nil {
		return tracking, err
	}
	if _, err = shared.WaitNativeFinality(ctx, confirmation, hash, 2*time.Minute); err != nil {
		return tracking, err
	}
	result, err := source.GetTracking(ctx, tracking.ID)
	if err != nil {
		return tracking, err
	}
	return provenance.TrackingFromResult(result, tracking.ID) // Pending remains pending.
}
