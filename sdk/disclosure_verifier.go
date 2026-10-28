package sdk

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/internal/receipts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type DisclosureReference struct {
	TransactionHash common.Hash
	TreeID          uint64
	LeafIndex       uint32
	Role            string
}
type DisclosureVerifierOptions struct {
	Protocol      *Protocol
	TokenAddress  common.Address
	Provider      Provider
	Confirmations uint64
}
type DisclosureVerifier struct{ options DisclosureVerifierOptions }
type DisclosureEvidence struct {
	Reference     DisclosureReference
	Block         protocolrpc.CanonicalBlock
	Confirmations uint64
	Output        DisclosureOutput
}
type VerifiedDisclosure struct {
	Result        DisclosureResult
	Block         protocolrpc.CanonicalBlock
	Confirmations uint64
	Reference     DisclosureReference
	Commitment    common.Hash
}

func NewDisclosureVerifier(options DisclosureVerifierOptions) (*DisclosureVerifier, error) {
	if options.Protocol == nil || options.Provider == nil || options.TokenAddress == (common.Address{}) {
		return nil, errors.New("protocol, token and Chain RPC provider required")
	}
	if options.Confirmations == 0 {
		options.Confirmations = 1
	}
	return &DisclosureVerifier{options}, nil
}
func (v *DisclosureVerifier) barrier(ctx context.Context, block protocolrpc.CanonicalBlock) (uint64, error) {
	if e := v.options.Protocol.ValidateDeployment(ctx, v.options.Provider); e != nil {
		return 0, e
	}
	header, e := v.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return 0, e
	}
	if header == nil || header.Hash() != block.Hash {
		return 0, errors.New("disclosure block is no longer canonical")
	}
	head, e := v.options.Provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return 0, e
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() || head.Number.Uint64() < block.Number || head.Number.Uint64()-block.Number+1 < v.options.Confirmations {
		return 0, errors.New("insufficient disclosure confirmations")
	}
	return head.Number.Uint64() - block.Number + 1, nil
}
func (v *DisclosureVerifier) ReadOutput(ctx context.Context, reference DisclosureReference) (DisclosureEvidence, error) {
	var result DisclosureEvidence
	if reference.TreeID == 0 || reference.LeafIndex >= 1<<core.TreeDepth || (reference.Role != "funds" && reference.Role != "owner" && reference.Role != "spender") {
		return result, errors.New("invalid disclosure reference")
	}
	if e := v.options.Protocol.ValidateDeployment(ctx, v.options.Provider); e != nil {
		return result, e
	}
	context := v.options.Protocol.Context()
	receipt, e := v.options.Provider.TransactionReceipt(ctx, reference.TransactionHash)
	if e != nil {
		return result, e
	}
	if receipt == nil || receipt.TxHash != reference.TransactionHash || receipt.Status != types.ReceiptStatusSuccessful || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		return result, errors.New("disclosure requires a successful receipt")
	}
	header, e := v.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return result, e
	}
	if header == nil || header.Hash() != receipt.BlockHash {
		return result, errors.New("noncanonical disclosure receipt")
	}
	logs := []types.Log{}
	for _, log := range receipt.Logs {
		if log.Address != context.ProtocolAddress {
			continue
		}
		if log.Removed || log.TxHash != reference.TransactionHash || log.BlockHash != receipt.BlockHash || log.BlockNumber != receipt.BlockNumber.Uint64() {
			return result, errors.New("invalid receipt log binding")
		}
		logs = append(logs, *log)
	}
	decoder, e := receipts.NewDecoder()
	if e != nil {
		return result, e
	}
	batch, e := decoder.Decode(receipts.Block{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash, ParentHash: header.ParentHash}, logs)
	if e != nil {
		return result, e
	}
	var selected *receipts.Output
	for i, output := range batch.Outputs {
		if output.TreeID == reference.TreeID && output.LeafIndex == reference.LeafIndex && output.Token == v.options.TokenAddress && output.Role == reference.Role {
			if selected != nil {
				return result, errors.New("ambiguous disclosure output")
			}
			selected = &batch.Outputs[i]
		}
	}
	if selected == nil {
		return result, errors.New("disclosure output not found in receipt")
	}
	block := protocolrpc.CanonicalBlock{Number: header.Number.Uint64(), Hash: header.Hash(), ParentHash: header.ParentHash, Timestamp: header.Time}
	confirmations, e := v.barrier(ctx, block)
	if e != nil {
		return result, e
	}
	return DisclosureEvidence{reference, block, confirmations, DisclosureOutput{core.ProtocolContext{ChainID: core.ElementFromUint64(context.ChainID), ProtocolAddress: core.BigIntToElement(context.ProtocolAddress.Big())}, core.BigIntToElement(v.options.TokenAddress.Big()), selected.Commitment, reference.Role, append([]byte(nil), selected.Ciphertext...)}}, nil
}
func (v *DisclosureVerifier) Verify(ctx context.Context, disclosure *Disclosure, reference DisclosureReference) (VerifiedDisclosure, error) {
	if disclosure == nil {
		return VerifiedDisclosure{}, errors.New("disclosure required")
	}
	evidence, e := v.ReadOutput(ctx, reference)
	if e != nil {
		return VerifiedDisclosure{}, e
	}
	result, e := disclosure.Verify(evidence.Output)
	if e != nil {
		return VerifiedDisclosure{}, e
	}
	confirmations, e := v.barrier(ctx, evidence.Block)
	if e != nil {
		return VerifiedDisclosure{}, e
	}
	return VerifiedDisclosure{result, evidence.Block, confirmations, reference, evidence.Output.OutputCommitment}, nil
}
