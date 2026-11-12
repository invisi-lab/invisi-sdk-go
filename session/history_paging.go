package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *WalletSession) recordStore() (RecordWalletStore, bool) {
	store := s.options.Store
	if n, ok := store.(*notifyingWalletStore); ok {
		store = n.base
	}
	r, ok := store.(RecordWalletStore)
	return r, ok
}
func (s *WalletSession) metadataSnapshot(ctx context.Context) (WalletState, error) {
	if r, ok := s.recordStore(); ok {
		return r.ReadMetadata(ctx, s.scope)
	}
	return s.options.Store.ReadSnapshot(ctx, s.scope)
}
func (s *WalletSession) fullHistorySnapshot(ctx context.Context) (WalletState, error) {
	v, e := s.fencedSnapshot(ctx)
	if e != nil {
		return v, e
	}
	r, ok := s.recordStore()
	if !ok || v.Records.Version == 0 {
		return s.fenceHistoryState(ctx, v)
	}
	for _, kind := range []string{"notes", "credits", "operations", "credit-operations"} {
		q := RecordQuery{Kind: kind, Limit: 256}
		for {
			page, e := r.ReadRecords(ctx, s.scope, q)
			if e != nil {
				return v, e
			}
			if page.Revision != v.Revision {
				return v, errors.New("wallet changed during full history read")
			}
			for _, record := range page.Items {
				if e = ApplyWalletRecord(&v, record); e != nil {
					return v, e
				}
			}
			if page.NextCursor == "" {
				break
			}
			q.Cursor = page.NextCursor
		}
	}
	end, e := r.ReadMetadata(ctx, s.scope)
	if e != nil {
		return v, e
	}
	if end.Revision != v.Revision {
		return v, errors.New("wallet changed during full history read")
	}
	return s.fenceHistoryState(ctx, v)
}

type HistoryPageOptions struct {
	Direction    string
	Limit        int
	Cursor       string
	TokenAddress *common.Address
}
type HistoryPageItem struct {
	Entry     any    `json:"entry"`
	Archived  bool   `json:"archived"`
	Canonical string `json:"canonical"`
}
type WalletHistoryPage struct {
	Items          []HistoryPageItem           `json:"items"`
	NextCursor     string                      `json:"nextCursor,omitempty"`
	Revision       uint64                      `json:"revision"`
	CheckedThrough *protocolrpc.CanonicalBlock `json:"checkedThrough"`
	Invalidated    bool                        `json:"invalidated"`
}

func (s *WalletSession) GetHistoryPage(ctx context.Context, o HistoryPageOptions) (WalletHistoryPage, error) {
	store, ok := s.recordStore()
	if !ok {
		return WalletHistoryPage{}, errors.New("wallet adapter does not support record paging")
	}
	kind := "notes"
	if o.Direction == "outgoing" {
		kind = "operations"
	} else if o.Direction != "incoming" {
		return WalletHistoryPage{}, errors.New("history direction must be incoming or outgoing")
	}
	summary, e := s.metadataSnapshot(ctx)
	if e != nil {
		return WalletHistoryPage{}, e
	}
	summary, e = s.fenceHistoryState(ctx, summary)
	if e != nil {
		return WalletHistoryPage{}, e
	}
	p, e := store.ReadRecords(ctx, s.scope, RecordQuery{Kind: kind, Limit: o.Limit, Cursor: o.Cursor})
	if e != nil {
		return WalletHistoryPage{}, e
	}
	end, e := s.metadataSnapshot(ctx)
	if e != nil {
		return WalletHistoryPage{}, e
	}
	if p.Revision != summary.Revision || end.Revision != summary.Revision {
		return WalletHistoryPage{}, errors.New("wallet history changed; restart pagination")
	}
	summary.Invalidated = summary.Invalidated || end.Invalidated
	summary, e = s.fenceHistoryState(ctx, summary)
	if e != nil {
		return WalletHistoryPage{}, e
	}
	out := WalletHistoryPage{Items: []HistoryPageItem{}, NextCursor: p.NextCursor, Revision: p.Revision, CheckedThrough: summary.DiscoveryThrough, Invalidated: summary.Invalidated}
	for _, r := range p.Items {
		canonical := "current"
		if summary.Invalidated || r.Generation != summary.Scan.Generation {
			canonical = "unknown"
		}
		v := emptyState()
		if e = ApplyWalletRecord(&v, r); e != nil {
			return out, e
		}
		var token common.Address
		if o.TokenAddress != nil {
			token = *o.TokenAddress
		}
		h, e := walletHistoryFromState(v, s.scope, o.TokenAddress != nil, token)
		if e != nil {
			return out, e
		}
		for _, entry := range h.Incoming {
			out.Items = append(out.Items, HistoryPageItem{Entry: entry, Archived: r.Archived, Canonical: canonical})
		}
		for _, entry := range h.Outgoing {
			out.Items = append(out.Items, HistoryPageItem{Entry: entry, Archived: r.Archived, Canonical: canonical})
		}
	}
	return out, nil
}

func (s *WalletSession) fenceHistoryState(ctx context.Context, state WalletState) (WalletState, error) {
	if s.publicHistory == nil {
		return state, ctx.Err()
	}
	public, err := s.publicHistory.Snapshot(ctx)
	if err != nil {
		return state, err
	}
	if public.Generation != state.Scan.Generation {
		state.Invalidated = true
	}
	return state, nil
}
