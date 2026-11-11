package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/chain"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"math/big"
	"sync"
)

type PrivateScanState struct {
	Generation                                                   uint64
	Coverage                                                     []CoverageRange
	RecoveryBoundary                                             uint64
	ChainHead, IndexedHead                                       *protocolrpc.CanonicalBlock
	CreditReplayTarget, CreditReplayCursor, CreditReplayRevision uint64
	Unresolved                                                   map[string]UnresolvedDelivery
}
type UnresolvedDelivery struct {
	Method      wallet.Method
	Params      wallet.DecryptParams
	Transaction PublicTransaction
	Anchor      protocolrpc.CanonicalBlock
}

var sharedHistories = struct {
	sync.Mutex
	values map[*sdk.Invisi]map[*chain.Protocol]*PublicHistory
}{values: map[*sdk.Invisi]map[*chain.Protocol]*PublicHistory{}}

func sharedPublicHistory(i *sdk.Invisi, p *chain.Protocol) (*PublicHistory, error) {
	sharedHistories.Lock()
	defer sharedHistories.Unlock()
	if h := sharedHistories.values[i][p]; h != nil {
		return h, nil
	}
	o := i.GetInfrastructure()
	h, e := NewPublicHistory(PublicHistoryOptions{Protocol: p, Provider: o.Provider, ProtocolRPC: o.ProtocolRPC})
	if e != nil {
		return nil, e
	}
	if sharedHistories.values[i] == nil {
		sharedHistories.values[i] = map[*chain.Protocol]*PublicHistory{}
	}
	sharedHistories.values[i][p] = h
	return h, nil
}

type HistoryProgress struct {
	Phase, Direction string
	Range            BlockRange
	Target           protocolrpc.CanonicalBlock
}
type ScanOptions struct {
	FromBlock, ToBlock               *uint64
	PageSize                         uint64
	Offline, Reset, SkipStateRefresh bool
	OnProgress                       func(HistoryProgress)
}

func (s *WalletSession) historyTarget(ctx context.Context, o ScanOptions) (protocolrpc.CanonicalBlock, *protocolrpc.CanonicalBlock, error) {
	status, e := s.rpc.GetStatus(ctx, s.protocol.Context())
	if e != nil {
		return protocolrpc.CanonicalBlock{}, nil, e
	}
	if status.IndexedHead == nil {
		return protocolrpc.CanonicalBlock{}, nil, errors.New("Protocol RPC has not indexed deployment")
	}
	target := *status.IndexedHead
	if o.ToBlock != nil {
		if *o.ToBlock > target.Number {
			return target, nil, errors.New("requested range exceeds indexed head")
		}
		h, e := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(*o.ToBlock))
		if e != nil {
			return target, nil, e
		}
		if h == nil || h.Number == nil {
			return target, nil, errors.New("invalid indexed target")
		}
		target = protocolrpc.CanonicalBlock{Number: h.Number.Uint64(), Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	}
	if e = s.canonical(ctx, target); e != nil {
		return target, nil, e
	}
	h, e := s.options.Provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return target, nil, e
	}
	if h == nil || h.Number == nil {
		return target, nil, errors.New("invalid chain head")
	}
	head := &protocolrpc.CanonicalBlock{Number: h.Number.Uint64(), Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	e = s.publicHistory.options.Store.Update(ctx, s.publicHistory.key, func(v *PublicHistoryState) error {
		if v.KnownHead == nil || v.KnownHead.Number <= head.Number {
			b := *head
			v.KnownHead = &b
		}
		if v.IndexedHead == nil || v.IndexedHead.Number <= target.Number {
			b := target
			v.IndexedHead = &b
		}
		return nil
	})
	if e != nil {
		return target, nil, e
	}
	return target, head, nil
}
func scanSize(o ScanOptions) (uint64, error) {
	n := o.PageSize
	if n == 0 {
		n = 100
	}
	if n > 256 {
		return 0, errors.New("pageSize must be 1..256")
	}
	return n, nil
}
func (s *WalletSession) SynchronizeHistory(ctx context.Context, o ScanOptions) error {
	size, e := scanSize(o)
	if e != nil {
		return e
	}
	target, _, e := s.historyTarget(ctx, o)
	if e != nil {
		return e
	}
	v, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return e
	}
	from := s.protocol.Deployment().Protocol.FromBlock
	if o.FromBlock != nil {
		from = *o.FromBlock
	}
	var ranges []BlockRange
	if o.FromBlock != nil {
		for n := from; n <= target.Number; {
			end := n + size - 1
			if end < n || end > target.Number {
				end = target.Number
			}
			ranges = append(ranges, BlockRange{n, end})
			if end == target.Number {
				break
			}
			n = end + 1
		}
	} else {
		ranges = PriorityRanges(v.Coverage, from, target.Number, size)
	}
	for i, q := range ranges {
		direction := "backward"
		var highest uint64
		for _, r := range v.Coverage {
			if r.Block.Number > highest {
				highest = r.Block.Number
			}
		}
		if o.FromBlock != nil || (len(v.Coverage) == 0 && i == 0) || (len(v.Coverage) > 0 && q.FromBlock > highest) {
			direction = "forward"
		}
		if o.OnProgress != nil {
			o.OnProgress(HistoryProgress{"processing", direction, q, target})
		}
		if _, e = s.publicHistory.Page(ctx, q, false); e != nil {
			return e
		}
		if o.OnProgress != nil {
			o.OnProgress(HistoryProgress{"committed", direction, q, target})
		}
	}
	return nil
}
func (s *WalletSession) ScanHistory(ctx context.Context, d wallet.DeliveryDecryptor, o ScanOptions) (RefreshResult, error) {
	select {
	case s.refreshGate <- struct{}{}:
		defer func() { <-s.refreshGate }()
	case <-ctx.Done():
		return RefreshResult{}, ctx.Err()
	}
	return s.scanHistory(ctx, d, o)
}

type stagedPublicPage struct {
	Page       PublicPage
	Generation uint64
}

func (s *WalletSession) scanHistory(ctx context.Context, d wallet.DeliveryDecryptor, o ScanOptions, staged ...*stagedPublicPage) (RefreshResult, error) {
	var result RefreshResult
	if d == nil {
		return result, errors.New("App delivery decryptor required")
	}
	size, e := scanSize(o)
	if e != nil {
		return result, e
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return result, e
	}
	public, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	target := protocolrpc.CanonicalBlock{}
	var head *protocolrpc.CanonicalBlock
	if o.Offline {
		if public.IndexedHead != nil {
			target = *public.IndexedHead
		} else if state.Scan.IndexedHead != nil {
			target = *state.Scan.IndexedHead
		} else {
			for _, r := range public.Coverage {
				if r.Block.Number > target.Number {
					target = r.Block
				}
			}
		}
		if o.ToBlock != nil && target.Number != *o.ToBlock {
			return result, errors.New("offline target unavailable")
		}
	} else {
		target, head, e = s.historyTarget(ctx, o)
		if e != nil {
			return result, e
		}
		if e = s.publicHistory.checkCanonical(ctx, public); e != nil {
			_ = s.invalidate(ctx)
			if !o.Reset || !errors.Is(e, ErrPublicHistoryReorg) {
				return result, e
			}
			public, e = s.publicHistory.Snapshot(ctx)
			if e != nil {
				return result, e
			}
		}
	}
	from := s.protocol.Deployment().Protocol.FromBlock
	if o.FromBlock != nil {
		from = *o.FromBlock
	}
	if from < s.protocol.Deployment().Protocol.FromBlock || from > target.Number {
		return result, errors.New("invalid history scan range")
	}
	if !o.Reset && !o.Offline {
		for _, saved := range []*protocolrpc.CanonicalBlock{state.ObservedThrough, state.CheckedThrough, state.DiscoveryThrough} {
			if saved != nil {
				if e = s.canonical(ctx, *saved); e != nil {
					_ = s.invalidate(ctx)
					return result, e
				}
			}
		}
	}
	if state.Scan.Generation != public.Generation || state.Invalidated {
		if !o.Reset {
			_ = s.invalidate(ctx)
			return result, errors.New("wallet requires reset scan")
		}
	}
	if o.Reset {
		if store, ok := s.recordStore(); ok {
			if restorer, ok := store.(JournalRestorer); ok {
				if e = restorer.RestoreExternalizedOperations(ctx, s.scope); e != nil {
					return result, e
				}
				state, e = s.options.Store.ReadSnapshot(ctx, s.scope)
				if e != nil {
					return result, e
				}
			}
		}
		for _, saved := range []*protocolrpc.CanonicalBlock{state.ObservedThrough, state.CheckedThrough, state.DiscoveryThrough} {
			if saved != nil && !o.Offline {
				if err := s.canonical(ctx, *saved); err != nil {
					if err = s.invalidate(ctx); err != nil {
						return result, err
					}
					state, e = s.options.Store.ReadSnapshot(ctx, s.scope)
					if e != nil {
						return result, e
					}
					break
				}
			}
		}
		if from != s.protocol.Deployment().Protocol.FromBlock {
			return result, errors.New("reset must start at deployment")
		}
		e = s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
			v.Notes = map[common.Hash]StoredNote{}
			v.PostpaidCredit.Grants = map[string]creditsession.StoredState{}
			v.Scan = PrivateScanState{Generation: public.Generation}
			v.DiscoveryThrough = nil
			v.ObservedThrough = nil
			v.CheckedThrough = nil
			v.Invalidated = true
			return nil
		})
		if e != nil {
			return result, e
		}
		state, e = s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			return result, e
		}
	}
	ranges := AccountPriorityRanges(state.Scan.Coverage, public.Pages, from, target.Number, size)
	var forwardStart *uint64
	if len(ranges) > 0 {
		if len(state.Scan.Coverage) == 0 {
			n := ranges[0].FromBlock
			forwardStart = &n
		} else {
			var highest uint64
			for _, r := range state.Scan.Coverage {
				if r.Block.Number > highest {
					highest = r.Block.Number
				}
			}
			if highest < target.Number {
				n := highest + 1
				forwardStart = &n
			}
		}
	}
	for _, q := range ranges {
		if e = ctx.Err(); e != nil {
			return result, e
		}
		direction := "backward"
		if forwardStart != nil && q.FromBlock >= *forwardStart {
			direction = "forward"
		}
		if o.OnProgress != nil {
			o.OnProgress(HistoryProgress{"processing", direction, q, target})
		}
		var page PublicPage
		var e error
		if len(staged) > 0 && staged[0] != nil && staged[0].Generation == public.Generation && staged[0].Page.Range == q {
			page = staged[0].Page
			staged[0] = nil
		} else {
			page, e = s.publicHistory.Page(ctx, q, o.Offline)
		}
		if e != nil {
			return result, e
		}
		before, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			return result, e
		}
		next, e := CloneState(before)
		if e != nil {
			return result, e
		}
		if next.Scan.Unresolved == nil {
			next.Scan.Unresolved = map[string]UnresolvedDelivery{}
		}
		resolver, e := s.pageResolver(ctx, d, page)
		if e != nil {
			return result, e
		}
		for _, tx := range page.Transactions {
			wrapped := func(m wallet.Method, p wallet.DecryptParams) (any, error) {
				v, e := resolver(m, p)
				if errors.Is(e, wallet.ErrAuthenticationFailed) {
					next.Scan.Unresolved[walletFingerprint(m, p)] = UnresolvedDelivery{m, p, tx, page.Block}
					return nil, wallet.ErrNotRecipient
				}
				return v, e
			}
			ns, cs, _, e := s.readPublicTransaction(ctx, tx, page.Block, wrapped)
			if e != nil {
				return result, e
			}
			for _, n := range ns {
				old, ok := next.Notes[n.Commitment]
				if !ok {
					result.AddedNotes++
				}
				n.ReservedBy = old.ReservedBy
				next.Notes[n.Commitment] = n
			}
			for _, c := range cs {
				k := creditsession.GrantKey(c.Commitment, c.Role)
				old, ok := next.PostpaidCredit.Grants[k]
				if !ok {
					result.AddedCredits++
				}
				c.ReservedBy = old.ReservedBy
				next.PostpaidCredit.Grants[k] = c
			}
		}
		next.Scan.Coverage, e = MergeCoverage(next.Scan.Coverage, CoverageRange{q.FromBlock, page.Block})
		if e != nil {
			return result, e
		}
		next.Scan.Generation = public.Generation
		next.Scan.IndexedHead = &target
		if head != nil {
			next.Scan.ChainHead = head
		}
		if next.Scan.RecoveryBoundary == 0 {
			next.Scan.RecoveryBoundary = q.FromBlock
		}
		creditWork := false
		for _, tx := range page.Transactions {
			if len(tx.Batch.CreditStates) > 0 || len(tx.Batch.CreditNullifiers) > 0 {
				creditWork = true
			}
		}
		if creditWork {
			if next.Scan.CreditReplayTarget < target.Number {
				next.Scan.CreditReplayTarget = target.Number
			}
			next.Scan.CreditReplayRevision++
			next.Scan.CreditReplayCursor = 0
		}
		next.ObservedThrough = &target
		next.DiscoveryThrough = nil
		if full := FullyScannedHeight(next.Scan.Coverage, s.protocol.Deployment().Protocol.FromBlock); full != nil {
			for _, r := range next.Scan.Coverage {
				if r.Block.Number == *full {
					b := r.Block
					next.DiscoveryThrough = &b
				}
			}
		}
		next.Invalidated = false
		if !o.Offline && !o.SkipStateRefresh {
			if e = s.checkKnownState(ctx, &next, *head); e != nil {
				return result, e
			}
		}
		if e = restoreReservations(&next); e != nil {
			return result, e
		}
		if e = restoreCreditReservations(&next.PostpaidCredit); e != nil {
			return result, e
		}
		if !o.Offline {
			if e = s.canonical(ctx, target); e != nil {
				_ = s.invalidate(ctx)
				return result, e
			}
		}
		e = s.commitPrivateBatch(ctx, public.Generation, before.Revision, next)
		if e != nil {
			return result, e
		}
		result.Block = &page.Block
		state = next
		if o.OnProgress != nil {
			o.OnProgress(HistoryProgress{"committed", direction, q, target})
		}
	}
	if !o.Offline {
		if e = s.replayCredits(ctx, target, size, o, &result); e != nil {
			return result, e
		}
		if !o.SkipStateRefresh {
			if e = s.RefreshState(ctx); e != nil {
				return result, e
			}
		}
	}
	return result, nil
}
func (s *WalletSession) deliveryResolver(ctx context.Context, d wallet.DeliveryDecryptor) walletResolver {
	return func(m wallet.Method, p wallet.DecryptParams) (any, error) {
		c, e := p.Context.ProtocolContext.Core()
		if e != nil {
			return nil, e
		}
		token, e := wallet.Element(p.Context.TokenAddress)
		if e != nil {
			return nil, e
		}
		commitment, e := wallet.Element(common.HexToHash(p.Context.OutputCommitment).Big().String())
		if e != nil {
			return nil, e
		}
		envelope, e := wallet.Bytes(p.Envelope, -1)
		if e != nil {
			return nil, e
		}
		if m == wallet.DecryptNote {
			return s.wallet.DecryptNote(ctx, d, envelope, primitives.NoteDecryptionContext{NoteEncryptionContext: primitives.NoteEncryptionContext{ProtocolContext: c, OutputCommitment: commitment}, TokenAddress: token, OwnerAddress: primitives.BigIntToElement(s.wallet.OwnerAddress().Big()), Clue: decodeClue(p.Context.Clue)})
		}
		role := credit.Owner
		if p.Context.Role == "spender" {
			role = credit.Spender
		}
		return s.wallet.DecryptCreditState(ctx, d, envelope, credit.DeliveryContext{ProtocolContext: c, TokenAddress: token, Commitment: commitment, RecipientIdentity: primitives.BigIntToElement(s.wallet.OwnerAddress().Big()), Clue: decodeClue(p.Context.Clue)}, role)
	}
}

func (s *WalletSession) stageSharedRefresh(ctx context.Context, o RefreshOptions) (*PendingWalletRefresh, error) {
	options := ScanOptions{FromBlock: refreshLowerBound(o.FromBlock), ToBlock: o.ToBlock, PageSize: o.PageSize, Reset: o.Reset}
	size, e := scanSize(options)
	if e != nil {
		return nil, e
	}
	target, _, e := s.historyTarget(ctx, options)
	if e != nil {
		return nil, e
	}
	base, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return nil, e
	}
	coverage := base.Scan.Coverage
	if o.Reset {
		coverage = nil
	}
	from := s.protocol.Deployment().Protocol.FromBlock
	if options.FromBlock != nil {
		from = *options.FromBlock
	}
	if from < s.protocol.Deployment().Protocol.FromBlock || from > target.Number {
		return nil, errors.New("invalid refresh range")
	}
	public, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	ranges := AccountPriorityRanges(coverage, public.Pages, from, target.Number, size)
	var staged *stagedPublicPage
	if len(ranges) > 0 {
		generation := public.Generation
		var page PublicPage
		if page, e = s.publicHistory.Page(ctx, ranges[0], false); e != nil {
			if o.Reset && errors.Is(e, ErrPublicHistoryReorg) {
				snapshot, err := s.publicHistory.Snapshot(ctx)
				if err != nil {
					return nil, err
				}
				generation = snapshot.Generation
				page, e = s.publicHistory.Page(ctx, ranges[0], false)
			}
			if e != nil {
				return nil, e
			}
		}
		public, e = s.publicHistory.Snapshot(ctx)
		if e != nil {
			return nil, e
		}
		if public.Generation != generation {
			return nil, errors.New("public history generation changed during staging")
		}
		staged = &stagedPublicPage{Page: page, Generation: generation}
	}
	height := target.Number
	o.ToBlock = &height
	return &PendingWalletRefresh{Block: &target, session: s, base: base, options: o, staged: staged}, nil
}

func refreshLowerBound(n uint64) *uint64 {
	if n == 0 {
		return nil
	}
	return &n
}
