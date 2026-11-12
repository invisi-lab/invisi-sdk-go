package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

type FinalityTag = protocol.FinalityTag
type FinalityStatus = protocol.FinalityStatus
type FinalityHead = protocol.FinalityHead
type ChainFinality = protocol.ChainFinality
type TransactionConfirmation = protocol.TransactionConfirmation

const (
	FinalityLatest       = protocol.FinalityLatest
	FinalitySafe         = protocol.FinalitySafe
	FinalityFinalized    = protocol.FinalityFinalized
	FinalityAvailable    = protocol.FinalityAvailable
	FinalityUnsupported  = protocol.FinalityUnsupported
	FinalityUnavailable  = protocol.FinalityUnavailable
	FinalityInconsistent = protocol.FinalityInconsistent
)

var errNativeFinalityInconsistent = errors.New("inconsistent native finality response")
var ErrFinalityUnsupported = errors.New("native finality tag unsupported")

type HeaderProvider interface {
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}
type TransactionConfirmationProvider interface {
	HeaderProvider
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
}

func finalityUnsupported(err error) bool {
	if errors.Is(err, ErrFinalityUnsupported) {
		return true
	}
	var rpcErr gethrpc.Error
	if !errors.As(err, &rpcErr) {
		return false
	}
	return rpcErr.ErrorCode() == -32601 || rpcErr.ErrorCode() == -32602
}
func nativeHeader(ctx context.Context, p HeaderProvider, tag FinalityTag) (*types.Header, error) {
	// ethclient omits the declared wire hash when decoding headers; verify it before promotion.
	if rawClient, ok := p.(interface{ Client() *gethrpc.Client }); ok {
		var raw json.RawMessage
		if err := rawClient.Client().CallContext(ctx, &raw, "eth_getBlockByNumber", string(tag), false); err != nil {
			return nil, err
		}
		var h *types.Header
		if err := json.Unmarshal(raw, &h); err != nil {
			return nil, errNativeFinalityInconsistent
		}
		if h == nil {
			return nil, nil
		}
		var wire struct {
			Hash common.Hash `json:"hash"`
		}
		if json.Unmarshal(raw, &wire) != nil || wire.Hash == (common.Hash{}) || wire.Hash != h.Hash() {
			return nil, errNativeFinalityInconsistent
		}
		return h, nil
	}
	if tagged, ok := p.(interface {
		HeaderByTag(context.Context, FinalityTag) (*types.Header, error)
	}); ok {
		return tagged.HeaderByTag(ctx, tag)
	}
	var n *big.Int
	switch tag {
	case FinalitySafe:
		n = big.NewInt(-4)
	case FinalityFinalized:
		n = big.NewInt(-3)
	}
	return p.HeaderByNumber(ctx, n)
}
func canonicalHeader(h *types.Header) (*CanonicalBlock, bool) {
	if h == nil || h.Number == nil || !h.Number.IsUint64() {
		return nil, false
	}
	return &CanonicalBlock{Number: h.Number.Uint64(), Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}, true
}
func readFinalityHead(ctx context.Context, p HeaderProvider, tag FinalityTag) FinalityHead {
	h, err := nativeHeader(ctx, p, tag)
	if err != nil {
		status := FinalityUnavailable
		if errors.Is(err, errNativeFinalityInconsistent) {
			status = FinalityInconsistent
		}
		if finalityUnsupported(err) {
			status = FinalityUnsupported
		}
		return FinalityHead{Status: status}
	}
	b, ok := canonicalHeader(h)
	if !ok {
		if h != nil {
			return FinalityHead{Status: FinalityInconsistent}
		}
		return FinalityHead{Status: FinalityUnavailable}
	}
	canonical, err := p.HeaderByNumber(ctx, new(big.Int).SetUint64(b.Number))
	if err != nil {
		return FinalityHead{Status: FinalityUnavailable}
	}
	actual, ok := canonicalHeader(canonical)
	if !ok || *actual != *b {
		return FinalityHead{Status: FinalityInconsistent}
	}
	return FinalityHead{Status: FinalityAvailable, Block: b}
}
func orderedFinality(f ChainFinality) bool {
	for _, pair := range [][2]FinalityHead{{f.Latest, f.Safe}, {f.Safe, f.Finalized}, {f.Latest, f.Finalized}} {
		a, b := pair[0].Block, pair[1].Block
		if a != nil && b != nil && (b.Number > a.Number || (a.Number == b.Number && *a != *b)) {
			return false
		}
	}
	return true
}
func ReadChainFinality(ctx context.Context, p HeaderProvider) ChainFinality {
	if p == nil {
		return ChainFinality{Latest: FinalityHead{Status: FinalityUnavailable}, Safe: FinalityHead{Status: FinalityUnavailable}, Finalized: FinalityHead{Status: FinalityUnavailable}}
	}
	for attempt := 0; attempt < 2; attempt++ {
		finalized := readFinalityHead(ctx, p, FinalityFinalized)
		safe := readFinalityHead(ctx, p, FinalitySafe)
		latest := readFinalityHead(ctx, p, FinalityLatest)
		f := ChainFinality{Latest: latest, Safe: safe, Finalized: finalized}
		ordered := orderedFinality(f)
		inconsistent := latest.Status == FinalityInconsistent || safe.Status == FinalityInconsistent || finalized.Status == FinalityInconsistent
		if ordered && !inconsistent {
			return f
		}
		if attempt == 1 {
			if !ordered {
				for _, head := range []*FinalityHead{&f.Latest, &f.Safe, &f.Finalized} {
					if head.Status == FinalityAvailable {
						*head = FinalityHead{Status: FinalityInconsistent}
					}
				}
			}
			return f
		}
	}
	panic("unreachable finality read")
}
func ReadTransactionConfirmation(ctx context.Context, p TransactionConfirmationProvider, hash common.Hash) TransactionConfirmation {
	pending := TransactionConfirmation{Level: "pending", Status: FinalityAvailable}
	if p == nil || hash == (common.Hash{}) {
		pending.Status = FinalityUnavailable
		return pending
	}
	receipt, err := p.TransactionReceipt(ctx, hash)
	if errors.Is(err, ethereum.NotFound) {
		return pending
	}
	if err != nil {
		pending.Status = FinalityUnavailable
		return pending
	}
	if receipt == nil {
		return pending
	}
	if receipt.TxHash != hash || receipt.Status > 1 || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		pending.Status = FinalityInconsistent
		return pending
	}
	h, err := p.HeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil {
		pending.Status = FinalityUnavailable
		return pending
	}
	b, ok := canonicalHeader(h)
	if !ok || b.Number != receipt.BlockNumber.Uint64() || b.Hash != receipt.BlockHash {
		pending.Status = FinalityInconsistent
		return pending
	}
	heads := ReadChainFinality(ctx, p)
	result := TransactionConfirmation{Level: "included", Status: heads.Finalized.Status, Block: b}
	for _, tag := range []FinalityTag{FinalitySafe, FinalityFinalized} {
		head := heads.Head(tag)
		if head.Status == FinalityAvailable && head.Block != nil && head.Block.Number >= b.Number {
			result.Level = string(tag)
		}
	}
	if result.Level == "finalized" {
		result.Status = FinalityAvailable
	}
	for _, head := range []FinalityHead{heads.Latest, heads.Safe, heads.Finalized} {
		if head.Status == FinalityInconsistent {
			result.Status = FinalityInconsistent
		}
	}
	h, err = p.HeaderByNumber(ctx, receipt.BlockNumber)
	again, ok := canonicalHeader(h)
	if err != nil {
		result.Status = FinalityUnavailable
		return result
	}
	if !ok || *again != *b {
		pending.Status = FinalityInconsistent
		return pending
	}
	return result
}

// RequireFinality never substitutes depth, elapsed time or indexer claims for a native tag.
func RequireFinality(ctx context.Context, p HeaderProvider, tag FinalityTag) (CanonicalBlock, error) {
	if !tag.Valid() {
		return CanonicalBlock{}, errors.New("invalid native finality tag")
	}
	head := ReadChainFinality(ctx, p).Head(tag)
	if head.Status != FinalityAvailable || head.Block == nil {
		return CanonicalBlock{}, errors.New("native finality " + string(head.Status))
	}
	return *head.Block, nil
}

func checkedFinality(ctx context.Context, p ChainProvider, raw json.RawMessage) (*ChainFinality, error) {
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil || len(value) != 3 {
		return nil, errors.New("invalid RPC finality")
	}
	claimed := ChainFinality{}
	for tag, head := range map[string]*FinalityHead{"latest": &claimed.Latest, "safe": &claimed.Safe, "finalized": &claimed.Finalized} {
		var fields map[string]json.RawMessage
		if json.Unmarshal(value[tag], &fields) != nil || len(fields) != 2 || fields["status"] == nil || fields["block"] == nil {
			return nil, errors.New("invalid RPC finality head")
		}
		if json.Unmarshal(value[tag], head) != nil || !head.Status.Valid() || (head.Status == FinalityAvailable) != (head.Block != nil) {
			return nil, errors.New("invalid RPC finality status/block")
		}
		if head.Block != nil {
			var blockFields map[string]json.RawMessage
			if json.Unmarshal(fields["block"], &blockFields) != nil || len(blockFields) != 4 || blockFields["number"] == nil || blockFields["hash"] == nil || blockFields["parentHash"] == nil || blockFields["timestamp"] == nil {
				return nil, errors.New("invalid RPC finality block fields")
			}
		}
		if head.Block != nil && (head.Block.Hash == (common.Hash{}) || head.Block.Number > 1<<53-1 || head.Block.Timestamp > 1<<53-1) {
			return nil, errors.New("invalid RPC finality block")
		}
	}
	if !orderedFinality(claimed) {
		return nil, errors.New("inconsistent RPC finality ordering")
	}
	own := ReadChainFinality(ctx, p)
	for tag, head := range map[FinalityTag]*FinalityHead{FinalityLatest: &claimed.Latest, FinalitySafe: &claimed.Safe, FinalityFinalized: &claimed.Finalized} {
		if head.Status != FinalityAvailable {
			continue
		}
		primary := own.Head(tag)
		if primary.Status != FinalityAvailable || primary.Block == nil {
			*head = FinalityHead{Status: primary.Status}
			continue
		}
		if head.Block.Number > primary.Block.Number {
			*head = FinalityHead{Status: FinalityUnavailable}
			continue
		}
		h, err := p.HeaderByNumber(ctx, new(big.Int).SetUint64(head.Block.Number))
		block, ok := canonicalHeader(h)
		if err != nil {
			*head = FinalityHead{Status: FinalityUnavailable}
		} else if !ok || *block != *head.Block {
			*head = FinalityHead{Status: FinalityInconsistent}
		}
	}
	return &claimed, nil
}
