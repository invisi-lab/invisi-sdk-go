package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

type NoteCandidate struct {
	Input  sdk.NoteInput
	Output rpc.ProtocolOutput
}
type CreditCandidate struct {
	Input  sdk.CreditInput
	Output rpc.ProtocolOutput
}

// DiscoverNotes scans genuine indexed outputs in 100-block pages, authenticates decryption,
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
		to := min(from+99, status.IndexedHead.Number)
		page, err := client.GetOutputs(ctx, deployment, rpc.ProtocolUpdateRange{FromBlock: from, ToBlock: to})
		if err != nil {
			return nil, err
		}
		for _, output := range page.Outputs {
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
			note, err := viewer.DecryptNote(output.Ciphertext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(d.ChainID), ProtocolAddress: core.BigIntToElement(d.ProtocolAddress.Big())}, OutputCommitment: core.BigIntToElement(output.Commitment.Big())}, TokenAddress: core.BigIntToElement(token.Big()), WalletIdentity: core.BigIntToElement(wallet.WalletIdentity().Big())})
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

// DiscoverCredits authenticates the requested role before deduplicating the two envelopes.
// Canonical receipt evidence also binds historical receiving-key snapshots and ciphertext.
// Expired Credit remains visible so its owner can select it for cancellation.
func DiscoverCredits(ctx context.Context, provider sdk.Provider, client *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.PublicWallet, viewingKey []byte, token common.Address, role string) ([]CreditCandidate, error) {
	if role != "owner" && role != "spender" {
		return nil, fmt.Errorf("role must be owner or spender")
	}
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
	verifier, err := sdk.NewDisclosureVerifier(sdk.DisclosureVerifierOptions{Protocol: pool, Provider: provider, TokenAddress: token, Confirmations: 1})
	if err != nil {
		return nil, err
	}
	seen := map[common.Hash]string{}
	positions := map[string]string{}
	result := []CreditCandidate{}
	for from := pool.Deployment().Protocol.FromBlock; from <= status.IndexedHead.Number; {
		to := min(from+99, status.IndexedHead.Number)
		page, err := client.GetCreditOutputs(ctx, deployment, rpc.ProtocolUpdateRange{FromBlock: from, ToBlock: to})
		if err != nil {
			return nil, err
		}
		for _, output := range page.Outputs {
			if output.TokenAddress != token {
				continue
			}
			state, err := viewer.DecryptCredit(output.Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(d.ChainID), ProtocolAddress: core.BigIntToElement(d.ProtocolAddress.Big())}, OutputCommitment: core.BigIntToElement(output.Commitment.Big())}, TokenAddress: core.BigIntToElement(token.Big()), WalletIdentity: core.BigIntToElement(wallet.WalletIdentity().Big())}, Role: core.CreditDeliveryRole(role)})
			if err != nil {
				continue
			}
			encoded, _ := json.Marshal(output)
			fingerprint := string(encoded)
			position := fmt.Sprintf("%d:%d", output.TreeID, output.LeafIndex)
			if prior, ok := seen[output.Commitment]; ok {
				if prior != fingerprint {
					return nil, fmt.Errorf("conflicting Credit role record")
				}
				continue
			}
			if prior, ok := positions[position]; ok && prior != fingerprint {
				return nil, fmt.Errorf("conflicting Credit position")
			}
			seen[output.Commitment] = fingerprint
			positions[position] = fingerprint
			nullifier, err := core.ComputeCreditNullifier(state.Note.ChainID, state.Note.CreditSharedKey, core.BigIntToElement(output.Commitment.Big()))
			if err != nil {
				return nil, err
			}
			spent, err := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(status.IndexedHead.Number)}, nullifier.Bytes())
			if err != nil {
				return nil, err
			}
			if spent {
				continue
			}
			evidence, err := verifier.ReadOutput(ctx, sdk.DisclosureReference{TransactionHash: output.TransactionHash, TreeID: output.TreeID, LeafIndex: output.LeafIndex, Role: role})
			if err != nil {
				return nil, err
			}
			if evidence.Output.OutputCommitment != output.Commitment || !bytes.Equal(evidence.Output.Ciphertext, output.Ciphertext) || evidence.Block != output.Block {
				return nil, fmt.Errorf("indexed Credit delivery does not match canonical receipt")
			}
			result = append(result, CreditCandidate{sdk.CreditInput{State: state, Commitment: output.Commitment, TreeID: output.TreeID, LeafIndex: output.LeafIndex, Role: role}, output})
		}
		if to == status.IndexedHead.Number {
			break
		}
		from = to + 1
	}
	return result, nil
}
