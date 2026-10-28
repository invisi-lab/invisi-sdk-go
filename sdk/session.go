package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/internal/receipts"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type Provider interface {
	protocolrpc.ChainProvider
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
}
type InvisiOptions struct {
	GasBenchmarks   []GasBenchmarkProfile
	ProtocolRPC     *protocolrpc.ProtocolRpcClient
	Provider        Provider
	Signer          *bind.TransactOpts
	Store           WalletStore
	Prover          Prover
	FundingSigner   FundingSigner
	FundingVerifier FundingAuthorizationVerifier
}
type Invisi struct{ options InvisiOptions }

func NewInvisi(options InvisiOptions) *Invisi {
	options.GasBenchmarks = copyGasProfiles(options.GasBenchmarks)
	return &Invisi{options}
}

type ConnectionOptions struct {
	Wallet        *Wallet
	Protocol      *protocol.Protocol
	Namespace     string
	Viewing       *ViewingCapability
	Authorization WalletAuthorizationProvider
}
type WalletSession struct {
	sends             *sessionSendState
	options           InvisiOptions
	wallet            *Wallet
	protocol          *protocol.Protocol
	scope             WalletScope
	viewing           *ViewingCapability
	authorization     WalletAuthorizationProvider
	rpc               *protocolrpc.ProtocolRpcClient
	explicitInventory *WalletState
	changeAddress     string
}

func (i *Invisi) Connect(ctx context.Context, options ConnectionOptions) (*WalletSession, error) {
	if i.options.Store == nil {
		return nil, errors.New("store required for a managed wallet session")
	}
	return i.resolve(ctx, options)
}
func (i *Invisi) resolve(ctx context.Context, options ConnectionOptions) (*WalletSession, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if options.Wallet == nil || options.Protocol == nil || i.options.Provider == nil || i.options.ProtocolRPC == nil {
		return nil, errors.New("wallet, protocol, Chain RPC and Protocol RPC required")
	}
	if e := options.Protocol.ValidateDeployment(ctx, i.options.Provider); e != nil {
		return nil, e
	}
	scope := WalletScope{ChainID: options.Protocol.Context().ChainID, ProtocolAddress: options.Protocol.Context().ProtocolAddress, WalletIdentity: options.Wallet.identity, Namespace: options.Namespace}
	if scope.Namespace == "" {
		scope.Namespace = "default"
	}
	viewer := options.Viewing
	if viewer == nil {
		viewer, _ = options.Wallet.GetViewer()
	}
	if viewer != nil && viewer.identity != options.Wallet.identity {
		return nil, errors.New("viewing identity mismatch")
	}
	authority := options.Authorization
	if authority == nil {
		if local, err := options.Wallet.ExportAuthorization(); err == nil {
			authority = local
		}
	}
	if authority != nil && authority.WalletIdentity() != options.Wallet.identity {
		return nil, errors.New("authorization identity mismatch")
	}
	// Load and compare RPC pins even when the Protocol was constructed from a private deployment record.
	d, e := i.options.ProtocolRPC.GetDeployment(ctx, scope.ChainID)
	if e != nil {
		return nil, e
	}
	expectedProfile, _ := json.Marshal(options.Protocol.Deployment())
	actualProfile, _ := json.Marshal(d)
	if string(expectedProfile) != string(actualProfile) {
		return nil, errors.New("Protocol RPC deployment mismatch")
	}
	return &WalletSession{sends: &sessionSendState{sending: map[string]bool{}}, options: i.options, wallet: options.Wallet, protocol: options.Protocol, scope: scope, viewing: viewer, authorization: authority, rpc: i.options.ProtocolRPC.WithProvider(i.options.Provider)}, nil
}
func (s *WalletSession) Scope() WalletScope           { return s.scope }
func (s *WalletSession) Protocol() *protocol.Protocol { return s.protocol }
func (s *WalletSession) coreContext() core.ProtocolContext {
	return core.ProtocolContext{ChainID: core.ElementFromUint64(s.scope.ChainID), ProtocolAddress: core.BigIntToElement(s.scope.ProtocolAddress.Big())}
}
func (s *WalletSession) canonical(ctx context.Context, block protocolrpc.CanonicalBlock) error {
	h, e := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return e
	}
	if h == nil || h.Number == nil || h.Number.Uint64() != block.Number || h.Hash() != block.Hash || h.ParentHash != block.ParentHash {
		return errors.New("canonical wallet checkpoint changed")
	}
	return nil
}
func (s *WalletSession) invalidate(ctx context.Context) error {
	return s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		state.Invalidated = true
		for id, record := range state.Operations {
			if (record.Status == "failed" || record.Status == "expired") && !record.Externalized {
				continue
			}
			record.Status = "reorged"
			record.ResolvedAt = nil
			record.WorkerID = ""
			state.Operations[id] = record
		}
		return nil
	})
}

type RefreshOptions struct {
	FromBlock    uint64
	ToBlock      *uint64
	DetectionKey []byte
	Reset        bool
	PageSize     uint64
}
type RefreshResult struct {
	Block                    *protocolrpc.CanonicalBlock
	AddedNotes, AddedCredits int
}

func (s *WalletSession) Refresh(ctx context.Context, options RefreshOptions) (RefreshResult, error) {
	var result RefreshResult
	if s.viewing == nil {
		return result, errors.New("viewing permission required")
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return result, e
	}
	if state.ObservedThrough != nil && !options.Reset {
		if e = s.canonical(ctx, *state.ObservedThrough); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
	}
	if state.CheckedThrough != nil && !options.Reset {
		if e = s.canonical(ctx, *state.CheckedThrough); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
	}
	if state.Invalidated && !options.Reset {
		return result, errors.New("wallet requires reset refresh")
	}
	status, e := s.rpc.GetStatus(ctx, s.protocol.Context())
	if e != nil {
		return result, e
	}
	if status.IndexedHead == nil {
		return result, errors.New("Protocol RPC has not indexed deployment")
	}
	end := status.IndexedHead.Number
	if options.ToBlock != nil {
		end = *options.ToBlock
	}
	if end > status.IndexedHead.Number {
		return result, errors.New("Protocol RPC is behind the requested range")
	}
	if options.FromBlock < s.protocol.Deployment().Protocol.FromBlock || options.FromBlock > end {
		return result, errors.New("invalid refresh range")
	}
	if !options.Reset && state.ObservedThrough != nil && end < state.ObservedThrough.Number {
		return result, errors.New("refresh anchor predates the saved checkpoint")
	}
	finalHeader, e := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(end))
	if e != nil {
		return result, e
	}
	if finalHeader == nil || finalHeader.Number == nil || finalHeader.Number.Uint64() != end {
		return result, errors.New("invalid refresh anchor")
	}
	finalBlock := protocolrpc.CanonicalBlock{Number: end, Hash: finalHeader.Hash(), ParentHash: finalHeader.ParentHash, Timestamp: finalHeader.Time}
	if !options.Reset && state.CheckedThrough != nil && end < state.CheckedThrough.Number {
		return result, errors.New("Protocol RPC is behind the latest checked spend state")
	}
	size := options.PageSize
	if size == 0 {
		size = 100
	}
	if size > 100 {
		return result, errors.New("pageSize must be 1..100")
	}
	if options.Reset {
		if e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
			state.Notes = map[common.Hash]StoredNote{}
			state.Credits = map[string]StoredCredit{}
			state.ObservedThrough = nil
			state.CheckedThrough = nil
			state.Invalidated = true
			return nil
		}); e != nil {
			return result, e
		}
	}
	for from := options.FromBlock; from <= end; {
		to := from + size - 1
		if to < from || to > end {
			to = end
		}
		page, e := s.rpc.GetUpdates(ctx, s.protocol.Context(), protocolrpc.ProtocolUpdateRange{FromBlock: from, ToBlock: to, DetectionKey: options.DetectionKey})
		if e != nil {
			return result, e
		}
		notes := []StoredNote{}
		credits := []StoredCredit{}
		draws := []*contracts.InvisiProtocolCreditDraw{}
		// Process complete receipts in canonical transaction order, independently of service ordering.
		type receiptOrder struct {
			hash  common.Hash
			block uint64
			index uint
		}
		order := []receiptOrder{}
		for _, hash := range page.TransactionHashes {
			receipt, e := s.options.Provider.TransactionReceipt(ctx, hash)
			if e != nil {
				return result, e
			}
			if receipt == nil || receipt.BlockNumber == nil || receipt.BlockNumber.Uint64() < from || receipt.BlockNumber.Uint64() > to {
				return result, errors.New("candidate transaction is outside the requested range")
			}
			order = append(order, receiptOrder{hash, receipt.BlockNumber.Uint64(), receipt.TransactionIndex})
		}
		sort.Slice(order, func(i, j int) bool {
			if order[i].block == order[j].block {
				return order[i].index < order[j].index
			}
			return order[i].block < order[j].block
		})
		page.TransactionHashes = nil
		for _, entry := range order {
			page.TransactionHashes = append(page.TransactionHashes, entry.hash)
		}
		for _, hash := range page.TransactionHashes {
			ns, cs, ds, e := s.readTransaction(ctx, hash, finalBlock)
			if e != nil {
				return result, e
			}
			notes = append(notes, ns...)
			credits = append(credits, cs...)
			draws = append(draws, ds...)
		}
		if e = s.canonical(ctx, finalBlock); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
		// Perform all network checks before the atomic local state update.
		before, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			return result, e
		}
		for _, c := range credits {
			before.Credits[creditKey(c.Commitment, c.Role)] = c
		}
		for _, draw := range draws {
			for _, previous := range before.Credits {
				for j, nf := range draw.Params.Nullifiers {
					if previous.Nullifier != common.Hash(nf) {
						continue
					}
					recovered, e := core.RecoverCreditRemainder(s.coreContext(), previous.State.Note, core.CreditRemainderRecovery{CreditNullifier: core.BigIntToElement(new(big.Int).SetBytes(nf[:])), RemainderCommitment: core.BigIntToElement(new(big.Int).SetBytes(draw.Params.RemainderCommitments[j][:])), RecoveryAmount: core.BigIntToElement(new(big.Int).SetBytes(draw.Params.RecoveryAmounts[j][:]))})
					if e != nil {
						return result, e
					}
					if recovered == nil {
						continue
					}
					commitment := common.Hash(draw.Params.RemainderCommitments[j])
					if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: draw.OutputTreeId.Uint64(), LeafIndex: draw.RemainderLeafIndices[j], Commitment: commitment}, finalBlock); e != nil {
						return result, e
					}
					child := previous
					child.State.Note = *recovered
					child.Commitment = commitment
					keyHash, _ := core.ComputeCreditSharedKeyHash(recovered.CreditSharedKey)
					leaf, e := core.ComputeCreditCommitment(recovered.ChainID, recovered.SpenderWalletIdentity, recovered.OwnerWalletIdentity, recovered.SpenderPayout, recovered.OwnerPayout, recovered.TokenAddress, recovered.Amount, recovered.Nonce, recovered.Expiry, keyHash)
					if e != nil {
						return result, e
					}
					cn, e := core.ComputeCreditNullifier(recovered.ChainID, recovered.CreditSharedKey, leaf)
					if e != nil {
						return result, e
					}
					child.Nullifier = fieldHash(cn)
					child.TreeID = draw.OutputTreeId.Uint64()
					child.LeafIndex = draw.RemainderLeafIndices[j]
					child.Spent = false
					child.ReservedBy = ""
					credits = append(credits, child)
					before.Credits[creditKey(child.Commitment, child.Role)] = child
				}
			}
		}
		for _, n := range notes {
			before.Notes[n.Commitment] = n
		}
		for _, c := range credits {
			before.Credits[creditKey(c.Commitment, c.Role)] = c
		}
		contract, e := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, s.options.Provider)
		if e != nil {
			return result, e
		}
		spentNotes := map[common.Hash]bool{}
		spentCredits := map[string]bool{}
		for hash, n := range before.Notes {
			spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(end)}, n.Nullifier)
			if e != nil {
				return result, e
			}
			spentNotes[hash] = spent
		}
		for key, c := range before.Credits {
			spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(end)}, c.Nullifier)
			if e != nil {
				return result, e
			}
			spentCredits[key] = spent
		}
		if e = s.canonical(ctx, finalBlock); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
		e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
			if (state.ObservedThrough != nil && state.ObservedThrough.Number > end) || (state.CheckedThrough != nil && state.CheckedThrough.Number > end) {
				return errors.New("concurrent refresh advanced beyond this anchor")
			}

			for _, n := range notes {
				old, ok := state.Notes[n.Commitment]
				if ok {
					n.ReservedBy = old.ReservedBy
				} else {
					result.AddedNotes++
				}
				state.Notes[n.Commitment] = n
			}
			for _, c := range credits {
				k := creditKey(c.Commitment, c.Role)
				old, ok := state.Credits[k]
				if ok {
					c.ReservedBy = old.ReservedBy
				} else {
					result.AddedCredits++
				}
				state.Credits[k] = c
			}
			for hash, spent := range spentNotes {
				n := state.Notes[hash]
				n.Spent = spent
				state.Notes[hash] = n
			}
			for key, spent := range spentCredits {
				c := state.Credits[key]
				c.Spent = spent
				state.Credits[key] = c
			}
			if state.ObservedThrough == nil || state.ObservedThrough.Number <= page.Block.Number {
				state.ObservedThrough = &page.Block
			}
			state.CheckedThrough = &finalBlock
			if err := restoreReservations(state); err != nil {
				return err
			}
			state.Invalidated = to != end && options.Reset
			return nil
		})
		if e != nil {
			return result, e
		}
		result.Block = &page.Block
		if to == end {
			break
		}
		from = to + 1
	}
	return result, nil
}
func creditKey(hash common.Hash, role string) string { return hash.Hex() + ":" + role }
func (s *WalletSession) readTransaction(ctx context.Context, hash common.Hash, anchor protocolrpc.CanonicalBlock) ([]StoredNote, []StoredCredit, []*contracts.InvisiProtocolCreditDraw, error) {
	receipt, e := s.options.Provider.TransactionReceipt(ctx, hash)
	if e != nil {
		return nil, nil, nil, e
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful || receipt.TxHash != hash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() || receipt.BlockNumber.Uint64() > anchor.Number {
		return nil, nil, nil, errors.New("invalid candidate receipt")
	}
	header, e := s.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return nil, nil, nil, e
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return nil, nil, nil, errors.New("candidate receipt is not canonical")
	}
	logs := []types.Log{}
	for _, l := range receipt.Logs {
		if l == nil {
			return nil, nil, nil, errors.New("nil receipt log")
		}
		if l.Address != s.scope.ProtocolAddress {
			continue
		}
		if l.Removed || l.TxHash != hash || l.BlockHash != receipt.BlockHash || l.BlockNumber != receipt.BlockNumber.Uint64() {
			return nil, nil, nil, errors.New("invalid receipt log binding")
		}
		logs = append(logs, *l)
	}
	decoder, e := receipts.NewDecoder()
	if e != nil {
		return nil, nil, nil, e
	}
	batch, e := decoder.Decode(receipts.Block{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash, ParentHash: header.ParentHash}, logs)
	if e != nil {
		return nil, nil, nil, e
	}
	notes := []StoredNote{}
	credits := []StoredCredit{}
	draws := []*contracts.InvisiProtocolCreditDraw{}
	filter, e := contracts.NewInvisiProtocolFilterer(s.scope.ProtocolAddress, nil)
	if e != nil {
		return nil, nil, nil, e
	}
	for _, l := range logs {
		if draw, e := filter.ParseCreditDraw(l); e == nil {
			draws = append(draws, draw)
		}
	}
	for _, output := range batch.Outputs {
		context := core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: s.coreContext(), OutputCommitment: core.BigIntToElement(output.Commitment.Big())}, TokenAddress: core.BigIntToElement(output.Token.Big()), WalletIdentity: core.BigIntToElement(s.wallet.identity.Big())}
		if note, e := s.viewing.decryptNote(output.Ciphertext, context); e == nil {
			if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: output.Commitment}, anchor); e != nil {
				return nil, nil, nil, e
			}
			nf, e := s.wallet.ExportNullifier().ForNote(note)
			if e != nil {
				return nil, nil, nil, e
			}
			notes = append(notes, StoredNote{Note: note, Commitment: output.Commitment, Nullifier: nf, TreeID: output.TreeID, LeafIndex: output.LeafIndex})
			continue
		}
		for _, role := range []core.CreditDeliveryRole{core.CreditOwner, core.CreditSpender} {
			state, e := s.viewing.decryptCredit(output.Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: context, Role: role})
			if e != nil {
				continue
			}
			if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: output.Commitment}, anchor); e != nil {
				return nil, nil, nil, e
			}
			nf, e := core.ComputeCreditNullifier(state.Note.ChainID, state.Note.CreditSharedKey, context.OutputCommitment)
			if e != nil {
				return nil, nil, nil, e
			}
			credits = append(credits, StoredCredit{State: state, Commitment: output.Commitment, Nullifier: fieldHash(nf), TreeID: output.TreeID, LeafIndex: output.LeafIndex, Role: string(role)})
		}
	}
	return notes, credits, draws, nil
}

type BalanceSnapshot struct {
	AvailableKnown, ConfirmedKnown, Reserved *big.Int
	TokenAddress                             common.Address
	Scope                                    WalletScope
	Invalidated                              bool
	CheckedThrough                           *protocolrpc.CanonicalBlock
	NoteCount                                int
}

func (s *WalletSession) Balance(ctx context.Context, token common.Address) (BalanceSnapshot, error) {
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return BalanceSnapshot{}, e
	}
	checked := state.CheckedThrough
	if checked == nil {
		checked = state.ObservedThrough
	}
	result := BalanceSnapshot{AvailableKnown: new(big.Int), ConfirmedKnown: new(big.Int), Reserved: new(big.Int), CheckedThrough: checked, TokenAddress: token, Scope: s.scope, Invalidated: state.Invalidated}
	for _, n := range state.Notes {
		if n.Spent || core.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
			continue
		}
		result.ConfirmedKnown.Add(result.ConfirmedKnown, core.ElementToBigInt(n.Note.Amount))
		if n.ReservedBy == "" {
			result.AvailableKnown.Add(result.AvailableKnown, core.ElementToBigInt(n.Note.Amount))
			result.NoteCount++
		} else {
			result.Reserved.Add(result.Reserved, core.ElementToBigInt(n.Note.Amount))
		}
	}
	if state.Invalidated || checked == nil {
		result.AvailableKnown = nil
	}
	if state.Invalidated {
		result.CheckedThrough = nil
	}
	return result, nil
}
func orderedNotes(state WalletState, token common.Address) []StoredNote {
	result := []StoredNote{}
	for _, n := range state.Notes {
		if !n.Spent && n.ReservedBy == "" && core.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) == 0 {
			result = append(result, n)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := core.ElementToBigInt(result[i].Note.Amount), core.ElementToBigInt(result[j].Note.Amount)
		if a.Cmp(b) == 0 {
			return result[i].Commitment.Hex() < result[j].Commitment.Hex()
		}
		return a.Cmp(b) < 0
	})
	return result
}
func (s *WalletSession) String() string {
	return fmt.Sprintf("WalletSession(%s,%s)", strconv.FormatUint(s.scope.ChainID, 10), s.wallet.identity.Hex())
}

// preparationInventory reads application inputs directly in explicit mode. It never implements or calls a WalletStore.
func (s *WalletSession) preparationInventory(ctx context.Context) (WalletState, error) {
	if err := ctx.Err(); err != nil {
		return WalletState{}, err
	}
	if s.explicitInventory != nil {
		return *s.explicitInventory, nil
	}
	return s.options.Store.ReadSnapshot(ctx, s.scope)
}
