package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"math/big"
)

type BlockProgress struct {
	FromBlock, ToBlock, TotalBlocks, ScannedBlocks uint64
	Complete                                       bool
	Ratio                                          *float64
	Unit                                           string
}
type ScanProgress struct{ Recent, Recovery, Overall BlockProgress }
type SuggestedScanRange struct {
	BlockRange
	Direction string
}
type WalletScanStatus struct {
	PublicThrough, AccountThrough                                    *protocolrpc.CanonicalBlock
	UnresolvedCount                                                  int
	AssetCompleteness                                                string
	Invalidated                                                      bool
	RecoveryStartHeight, RecoveryBoundary                            uint64
	ChainHeight, IndexedHeight, MaxScannedHeight, FullyScannedHeight *uint64
	ScanComplete, IsSynced, CreditRecoveryPending                    bool
	UnscannedRanges                                                  []BlockRange
	SuggestedScanRanges                                              []SuggestedScanRange
	NextAction                                                       string
	Progress                                                         *ScanProgress
}

func progress(r []CoverageRange, from, to uint64, empty bool) BlockProgress {
	p := BlockProgress{FromBlock: from, ToBlock: to, Unit: "blocks"}
	if !empty && to >= from {
		p.TotalBlocks = to - from + 1
		for _, v := range r {
			a, b := v.FromBlock, v.Block.Number
			if a < from {
				a = from
			}
			if b > to {
				b = to
			}
			if a <= b {
				p.ScannedBlocks += b - a + 1
			}
		}
		ratio := float64(new(big.Int).Div(new(big.Int).Mul(new(big.Int).SetUint64(p.ScannedBlocks), big.NewInt(10000)), new(big.Int).SetUint64(p.TotalBlocks)).Uint64()) / 10000
		p.Ratio = &ratio
	}
	p.Complete = p.ScannedBlocks == p.TotalBlocks
	return p
}
func BuildScanStatus(start uint64, v WalletState, public PublicHistoryState) WalletScanStatus {
	r := v.Scan.Coverage
	if len(r) == 0 && v.DiscoveryThrough != nil {
		r = []CoverageRange{{start, *v.DiscoveryThrough}}
	}
	if v.Invalidated {
		r = nil
	}
	s := WalletScanStatus{RecoveryStartHeight: start, RecoveryBoundary: v.Scan.RecoveryBoundary, Invalidated: v.Invalidated, UnresolvedCount: len(v.Scan.Unresolved), AssetCompleteness: "unknown", CreditRecoveryPending: v.Scan.CreditReplayTarget != 0, UnscannedRanges: []BlockRange{}, SuggestedScanRanges: []SuggestedScanRange{}}
	for _, c := range public.Coverage {
		if s.PublicThrough == nil || s.PublicThrough.Number < c.Block.Number {
			b := c.Block
			s.PublicThrough = &b
		}
	}
	if v.Records.Version == 1 {
		s.UnresolvedCount = int(v.Records.Counts["unresolved"])
	}
	indexed := public.IndexedHead
	if indexed == nil {
		indexed = v.Scan.IndexedHead
	}
	if indexed == nil {
		indexed = s.PublicThrough
	}
	head := public.KnownHead
	if head == nil {
		head = v.Scan.ChainHead
	}
	if head == nil {
		head = indexed
	}
	if indexed != nil {
		n := indexed.Number
		s.IndexedHeight = &n
	}
	if head != nil {
		n := head.Number
		s.ChainHeight = &n
	}
	for _, c := range r {
		if s.MaxScannedHeight == nil || *s.MaxScannedHeight < c.Block.Number {
			n := c.Block.Number
			s.MaxScannedHeight = &n
		}
	}
	s.FullyScannedHeight = FullyScannedHeight(r, start)
	if s.FullyScannedHeight != nil {
		for _, c := range r {
			if c.Block.Number == *s.FullyScannedHeight {
				b := c.Block
				s.AccountThrough = &b
			}
		}
	}
	s.ScanComplete = !v.Invalidated && head != nil && s.AccountThrough != nil && s.AccountThrough.Number == head.Number && s.AccountThrough.Hash == head.Hash
	s.IsSynced = s.ScanComplete && !s.CreditRecoveryPending
	if s.RecoveryBoundary == 0 {
		s.RecoveryBoundary = start
		if indexed != nil {
			pages := AccountPriorityRanges(r, public.Pages, start, indexed.Number, 100)
			if len(pages) > 0 {
				s.RecoveryBoundary = pages[0].FromBlock
			} else if s.FullyScannedHeight != nil {
				s.RecoveryBoundary = *s.FullyScannedHeight + 1
			}
		}
	}
	if head != nil {
		s.UnscannedRanges = MissingRanges(r, start, head.Number)
		recent := progress(r, s.RecoveryBoundary, head.Number, false)
		recovery := progress(r, start, s.RecoveryBoundary-1, s.RecoveryBoundary == 0)
		overall := progress(r, start, head.Number, false)
		s.Progress = &ScanProgress{recent, recovery, overall}
	}
	if indexed != nil {
		var forward *uint64
		if s.MaxScannedHeight == nil {
			pages := AccountPriorityRanges(r, public.Pages, start, indexed.Number, 100)
			if len(pages) > 0 {
				n := pages[0].FromBlock
				forward = &n
			}
		} else if *s.MaxScannedHeight < indexed.Number {
			n := *s.MaxScannedHeight + 1
			forward = &n
		}
		olderTo := indexed.Number
		haveOlder := true
		if forward != nil {
			s.SuggestedScanRanges = append(s.SuggestedScanRanges, SuggestedScanRange{BlockRange{*forward, indexed.Number}, "forward"})
			if *forward == 0 {
				haveOlder = false
			} else {
				olderTo = *forward - 1
			}
		}
		if haveOlder {
			gaps := MissingRanges(r, start, olderTo)
			for i := len(gaps) - 1; i >= 0; i-- {
				s.SuggestedScanRanges = append(s.SuggestedScanRanges, SuggestedScanRange{gaps[i], "backward"})
			}
		}
	}

	switch {
	case v.Invalidated:
		s.NextAction = "rebuild"
	case indexed == nil:
		s.NextAction = "discover-head"
	case s.MaxScannedHeight == nil || *s.MaxScannedHeight < indexed.Number:
		s.NextAction = "update-head"
	case len(s.SuggestedScanRanges) > 0:
		s.NextAction = "recover-history"
	case s.CreditRecoveryPending:
		s.NextAction = "recover-credit"
	case s.UnresolvedCount > 0:
		s.NextAction = "retry-unresolved"
	case !s.IsSynced:
		s.NextAction = "wait-for-indexer"
	default:
		s.NextAction = "none"
	}
	if s.IsSynced && s.UnresolvedCount == 0 {
		s.AssetCompleteness = "complete"
	}
	return s
}
func (s *WalletSession) GetScanStatus(ctx context.Context) (WalletScanStatus, error) {
	v, e := s.metadataSnapshot(ctx)
	if e != nil {
		return WalletScanStatus{}, e
	}
	p, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return WalletScanStatus{}, e
	}
	if v.Scan.Generation != p.Generation {
		v.Invalidated = true
	}
	return BuildScanStatus(s.protocol.Deployment().Protocol.FromBlock, v, p), nil
}

type WalletFundsStatus struct {
	TokenAddress                                                  common.Address
	SpendableKnown, ReservedKnown                                 *big.Int
	CheckedThrough                                                *protocolrpc.CanonicalBlock
	AreFundsSpendable, RequiresPreparationValidation, Invalidated bool
}

func (s *WalletSession) GetFundsStatus(ctx context.Context, token common.Address) (WalletFundsStatus, error) {
	v, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return WalletFundsStatus{}, e
	}
	out := WalletFundsStatus{TokenAddress: token, SpendableKnown: new(big.Int), ReservedKnown: new(big.Int), RequiresPreparationValidation: true, Invalidated: v.Invalidated}
	p, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return out, e
	}
	if p.Generation != v.Scan.Generation {
		out.Invalidated = true
	}
	head := p.KnownHead
	if head == nil {
		head = v.Scan.ChainHead
	}
	valid := !out.Invalidated && head != nil
	for _, n := range v.Notes {
		if common.BigToAddress(primitives.ElementToBigInt(n.Note.TokenAddress)) != token || n.Spent {
			continue
		}
		amount := primitives.ElementToBigInt(n.Note.Amount)
		if n.ReservedBy != "" {
			out.ReservedKnown.Add(out.ReservedKnown, amount)
			continue
		}
		if valid && n.CheckedThrough != nil && n.CheckedThrough.Number == head.Number && n.CheckedThrough.Hash == head.Hash {
			if out.CheckedThrough == nil {
				b := *n.CheckedThrough
				out.CheckedThrough = &b
			}
			out.SpendableKnown.Add(out.SpendableKnown, amount)
		}
	}
	out.AreFundsSpendable = out.SpendableKnown.Sign() > 0
	return out, nil
}
