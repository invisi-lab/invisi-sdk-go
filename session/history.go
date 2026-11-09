package session

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

// HistoryOptions selects an asset without exposing the internal persisted plan.
type HistoryOptions struct{ TokenAddress *common.Address }

type HistoryOutput struct {
	To      string
	Amount  *big.Int
	Payload []byte
	Change  bool
}

type HistoryFunding struct {
	Payer         common.Address
	Amount        *big.Int
	Authorization string
}

// OutgoingHistoryEntry contains business intent, not private witnesses or funding signatures.
type OutgoingHistoryEntry struct {
	ID, Kind, Status  string
	TokenAddress      common.Address
	TransactionHash   common.Hash
	Deadline          uint64
	PreparedAt        protocolrpc.CanonicalBlock
	Permissions       sdk.OperationPermissions
	InputCount        int
	Outputs           []HistoryOutput
	Funding           HistoryFunding
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
}

// IncomingHistoryEntry retains spent receipts; spending is not a failed incoming payment.
type IncomingHistoryEntry struct {
	TokenAddress    common.Address
	Amount          *big.Int
	Payload         []byte
	Commitment      common.Hash
	TreeID          uint64
	LeafIndex       uint32
	TransactionHash common.Hash
	Block           *protocolrpc.CanonicalBlock
	Spent           bool
	ReservedBy      string
}

type WalletHistory struct {
	Scope          WalletScope
	Revision       uint64
	CheckedThrough *protocolrpc.CanonicalBlock
	Invalidated    bool
	Outgoing       []OutgoingHistoryEntry
	Incoming       []IncomingHistoryEntry
}

// GetHistory returns detached snapshots. Reconciliation progress is not discovery coverage.
func (s *WalletSession) GetHistory(ctx context.Context, options HistoryOptions) (WalletHistory, error) {
	// Copy pointer-valued inputs before the first caller-owned Store operation.
	var tokenFilter common.Address
	filterToken := options.TokenAddress != nil
	if filterToken {
		tokenFilter = *options.TokenAddress
	}
	if s.options.Store == nil {
		return WalletHistory{}, errors.New("history requires a managed wallet store")
	}
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return WalletHistory{}, err
	}
	result := WalletHistory{Scope: s.scope, Revision: state.Revision, Invalidated: state.Invalidated,
		Outgoing: []OutgoingHistoryEntry{}, Incoming: []IncomingHistoryEntry{}}
	if state.DiscoveryThrough != nil {
		anchor := *state.DiscoveryThrough
		result.CheckedThrough = &anchor
	}
	for _, record := range state.Operations {
		var plan transaction.Plan
		if err := json.Unmarshal(record.Plan, &plan); err != nil {
			return WalletHistory{}, err
		}
		if filterToken && plan.Context.TokenAddress != tokenFilter {
			continue
		}
		operation := &Operation{plan: plan, status: record.Status}
		preview := operation.Preview()
		entry := OutgoingHistoryEntry{
			ID: record.ID, Kind: preview.Kind, Status: record.Status, TokenAddress: plan.Context.TokenAddress,
			TransactionHash: record.TransactionHash, Deadline: plan.Deadline, PreparedAt: plan.Block,
			Permissions: preview.Permissions, InputCount: preview.InputCount,
			Funding:           HistoryFunding{Payer: preview.FundingPayer, Amount: preview.FundingAmount, Authorization: plan.FundingMode},
			WithdrawalAddress: preview.WithdrawalAddress, WithdrawalAmount: preview.WithdrawalAmount,
			Outputs: []HistoryOutput{},
		}
		for _, output := range plan.Outputs {
			entry.Outputs = append(entry.Outputs, HistoryOutput{To: output.Address, Amount: new(big.Int).Set(output.Amount), Payload: outputPayload(output), Change: output.Change})
		}
		result.Outgoing = append(result.Outgoing, entry)
	}
	for _, record := range state.Notes {
		token := common.BigToAddress(core.ElementToBigInt(record.Note.TokenAddress))
		if filterToken && token != tokenFilter {
			continue
		}
		entry := IncomingHistoryEntry{TokenAddress: token, Amount: core.ElementToBigInt(record.Note.Amount),
			Payload: append([]byte(nil), record.Note.Payload...), Commitment: record.Commitment, TreeID: record.TreeID,
			LeafIndex: record.LeafIndex, TransactionHash: record.TransactionHash, Spent: record.Spent, ReservedBy: record.ReservedBy}
		if record.Block != nil {
			block := *record.Block
			entry.Block = &block
		}
		result.Incoming = append(result.Incoming, entry)
	}
	sort.Slice(result.Outgoing, func(i, j int) bool {
		if result.Outgoing[i].PreparedAt.Number != result.Outgoing[j].PreparedAt.Number {
			return result.Outgoing[i].PreparedAt.Number < result.Outgoing[j].PreparedAt.Number
		}
		return result.Outgoing[i].ID < result.Outgoing[j].ID
	})
	sort.Slice(result.Incoming, func(i, j int) bool {
		left, right := result.Incoming[i], result.Incoming[j]
		if left.TreeID != right.TreeID {
			return left.TreeID < right.TreeID
		}
		if left.LeafIndex != right.LeafIndex {
			return left.LeafIndex < right.LeafIndex
		}
		return left.Commitment.Hex() < right.Commitment.Hex()
	})
	return result, nil
}

type sessionNotifications struct {
	mu        sync.Mutex
	next      uint64
	listeners map[uint64]func()
	pending   [][]uint64
	running   bool
}

func newSessionNotifications() *sessionNotifications {
	return &sessionNotifications{listeners: map[uint64]func(){}}
}

func (n *sessionNotifications) subscribe(listener func()) func() {
	if n == nil || listener == nil {
		return func() {}
	}
	n.mu.Lock()
	id := n.next
	n.next++
	n.listeners[id] = listener
	n.mu.Unlock()
	return func() { n.mu.Lock(); delete(n.listeners, id); n.mu.Unlock() }
}

// notify queues callbacks after commit; callers may still hold an operation or refresh lock.
func (n *sessionNotifications) notify() {
	if n == nil {
		return
	}
	n.mu.Lock()
	ids := make([]uint64, 0, len(n.listeners))
	for id := range n.listeners {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		n.mu.Unlock()
		return
	}
	n.pending = append(n.pending, ids)
	if n.running {
		n.mu.Unlock()
		return
	}
	n.running = true
	n.mu.Unlock()
	go n.drain()
}

// One worker per active burst serializes queued callbacks without retaining an idle goroutine.
func (n *sessionNotifications) drain() {
	for {
		n.mu.Lock()
		if len(n.pending) == 0 {
			n.pending = nil
			n.running = false
			n.mu.Unlock()
			return
		}
		ids := n.pending[0]
		n.pending[0] = nil
		n.pending = n.pending[1:]
		n.mu.Unlock()
		for _, id := range ids {
			n.mu.Lock()
			listener := n.listeners[id]
			n.mu.Unlock()
			if listener != nil {
				func() { defer func() { _ = recover() }(); listener() }()
			}
		}
	}
}

// Subscribe asynchronously notifies this session's successful commits with serialized callbacks.
// Unsubscribe cancels queued callbacks; host applications own cross-process transport.
func (s *WalletSession) Subscribe(listener func()) func() { return s.history.subscribe(listener) }

type notifyingWalletStore struct {
	base          WalletStore
	notifications *sessionNotifications
}

func (s *notifyingWalletStore) ReadSnapshot(ctx context.Context, scope WalletScope) (WalletState, error) {
	return s.base.ReadSnapshot(ctx, scope)
}

func (s *notifyingWalletStore) Update(ctx context.Context, scope WalletScope, update func(*WalletState) error) error {
	if err := s.base.Update(ctx, scope, update); err != nil {
		return err
	}
	s.notifications.notify()
	return nil
}

func outputPayload(o transaction.PreparedOutput) []byte {
	if o.Note == nil {
		return nil
	}
	return append([]byte(nil), o.Note.Payload...)
}
