package session

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

type walletResolver func(wallet.Method, wallet.DecryptParams) (any, error)
type refreshPages struct {
	replay      bool
	notePages   map[string]protocolrpc.ProtocolNoteUpdates
	creditPages map[string]protocolrpc.ProtocolCreditUpdates
}

func rangeKey(q protocolrpc.ProtocolUpdateRange) string {
	return fmt.Sprintf("%d:%d", q.FromBlock, q.ToBlock)
}
func (p *refreshPages) notes(ctx context.Context, s *WalletSession, q protocolrpc.ProtocolUpdateRange) (protocolrpc.ProtocolNoteUpdates, error) {
	key := rangeKey(q)
	if p.replay {
		v, ok := p.notePages[key]
		if !ok {
			return v, errors.New("refresh candidate page changed")
		}
		return v, nil
	}
	v, err := s.rpc.GetNoteUpdates(ctx, s.protocol.Context(), q)
	if err == nil {
		p.notePages[key] = v
	}
	return v, err
}
func (p *refreshPages) credits(ctx context.Context, s *WalletSession, q protocolrpc.ProtocolUpdateRange) (protocolrpc.ProtocolCreditUpdates, error) {
	key := rangeKey(q)
	if p.replay {
		v, ok := p.creditPages[key]
		if !ok {
			return v, errors.New("refresh candidate page changed")
		}
		return v, nil
	}
	v, err := s.rpc.GetCreditTransitions(ctx, s.protocol.Context(), q)
	if err == nil {
		p.creditPages[key] = v
	}
	return v, err
}

// PendingWalletRefresh retains canonical candidate pages until App decryption is validated.
type PendingWalletRefresh struct {
	Block   *protocolrpc.CanonicalBlock
	mu      sync.Mutex
	applied bool
	session *WalletSession
	base    WalletState
	options RefreshOptions
	pages   *refreshPages
}

func (s *WalletSession) cloneForRefresh(base WalletState) (*WalletSession, error) {
	snapshot, err := CloneState(base)
	if err != nil {
		return nil, err
	}
	temp := *s
	temp.options = s.options
	temp.options.Store = &MemoryWalletStore{states: map[string]WalletState{s.scope.Key(): snapshot}}
	temp.refreshGate = make(chan struct{}, 1)
	return &temp, nil
}
func walletFingerprint(m wallet.Method, p wallet.DecryptParams) string {
	return string(m) + ":" + string(wallet.Encode(p))
}
func deliveryRequest(envelope []byte, c core.ProtocolContext, token string, commitment common.Hash, role string) wallet.DecryptParams {
	return wallet.DecryptParams{Envelope: wallet.Hex(envelope), Context: wallet.DeliveryContext{ProtocolContext: wallet.Context(c), TokenAddress: token, OutputCommitment: commitment.Hex(), Role: role}}
}
func resolveNote(resolvers []walletResolver, envelope []byte, c core.NoteDecryptionContext) (core.Note, error) {
	if len(resolvers) == 0 || resolvers[0] == nil {
		return core.Note{}, errors.New("wallet response required")
	}
	v, err := resolvers[0](wallet.DecryptNote, deliveryRequest(envelope, c.ProtocolContext, c.TokenAddress.String(), common.Hash(c.OutputCommitment.Bytes()), ""))
	if err != nil {
		return core.Note{}, err
	}
	note, ok := v.(core.Note)
	if !ok {
		return core.Note{}, errors.New("invalid note response")
	}
	return note, nil
}
func resolveCredit(resolvers []walletResolver, envelope []byte, c credit.DeliveryContext, role credit.Role) (credit.DeliveryState, error) {
	if len(resolvers) == 0 || resolvers[0] == nil {
		return credit.DeliveryState{}, errors.New("wallet response required")
	}
	name := "owner"
	if role == credit.Spender {
		name = "spender"
	}
	v, err := resolvers[0](wallet.DecryptCredit, deliveryRequest(envelope, c.ProtocolContext, c.TokenAddress.String(), common.Hash(c.Commitment.Bytes()), name))
	if err != nil {
		return credit.DeliveryState{}, err
	}
	state, ok := v.(credit.DeliveryState)
	if !ok {
		return credit.DeliveryState{}, errors.New("invalid credit response")
	}
	return state, nil
}

// Refresh stages canonical candidate discovery without changing the real inventory.
// Inject an App DeliveryDecryptor to validate and atomically apply the staged inventory.
func (s *WalletSession) Refresh(ctx context.Context, options RefreshOptions) (*PendingWalletRefresh, error) {
	select {
	case s.refreshGate <- struct{}{}:
		defer func() { <-s.refreshGate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	base, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return nil, err
	}
	if options.ToBlock != nil {
		height := *options.ToBlock
		options.ToBlock = &height
	}
	pending := &PendingWalletRefresh{session: s, base: base, options: options, pages: &refreshPages{notePages: map[string]protocolrpc.ProtocolNoteUpdates{}, creditPages: map[string]protocolrpc.ProtocolCreditUpdates{}}}
	record := func(wallet.Method, wallet.DecryptParams) (any, error) { return nil, wallet.ErrNotRecipient }
	temp, err := s.cloneForRefresh(base)
	if err != nil {
		return nil, err
	}
	result, err := temp.refreshInternal(ctx, options, record, pending.pages)
	if err != nil {
		return nil, err
	}
	if result.Block != nil {
		block := *result.Block
		pending.Block = &block
		height := block.Number
		pending.options.ToBlock = &height
	}
	return pending, nil
}
func (p *PendingWalletRefresh) Decrypt(ctx context.Context, decryptor wallet.DeliveryDecryptor) (RefreshResult, error) {
	if !p.mu.TryLock() {
		return RefreshResult{}, errors.New("wallet refresh apply already running")
	}
	defer p.mu.Unlock()
	if p.applied {
		return RefreshResult{}, errors.New("wallet refresh already applied")
	}
	s := p.session
	select {
	case s.refreshGate <- struct{}{}:
		defer func() { <-s.refreshGate }()
	case <-ctx.Done():
		return RefreshResult{}, ctx.Err()
	}
	current, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return RefreshResult{}, err
	}
	if current.Revision != p.base.Revision {
		return RefreshResult{}, errors.New("wallet inventory changed during pending refresh")
	}
	if decryptor == nil {
		return RefreshResult{}, errors.New("App delivery decryptor required")
	}
	replay := func(method wallet.Method, params wallet.DecryptParams) (any, error) {
		c, err := params.Context.ProtocolContext.Core()
		if err != nil {
			return nil, err
		}
		token, err := wallet.Element(params.Context.TokenAddress)
		if err != nil {
			return nil, err
		}
		commitment, err := wallet.Element(common.HexToHash(params.Context.OutputCommitment).Big().String())
		if err != nil {
			return nil, err
		}
		envelope, err := wallet.Bytes(params.Envelope, -1)
		if err != nil {
			return nil, err
		}
		if method == wallet.DecryptNote {
			return s.wallet.DecryptNote(ctx, decryptor, envelope, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: c, OutputCommitment: commitment}, TokenAddress: token})
		}
		role := credit.Owner
		if params.Context.Role == "spender" {
			role = credit.Spender
		}
		return s.wallet.DecryptCreditState(ctx, decryptor, envelope, credit.DeliveryContext{ProtocolContext: c, TokenAddress: token, Commitment: commitment}, role)
	}
	temp, err := s.cloneForRefresh(p.base)
	if err != nil {
		return RefreshResult{}, err
	}
	pages := *p.pages
	pages.replay = true
	result, err := temp.refreshInternal(ctx, p.options, replay, &pages)
	if err != nil {
		return RefreshResult{}, err
	}
	final, err := temp.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return RefreshResult{}, err
	}
	err = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		if state.Revision != p.base.Revision {
			return errors.New("wallet inventory changed before refresh apply")
		}
		final.Revision = state.Revision
		*state = final
		return nil
	})
	if err != nil {
		return RefreshResult{}, err
	}
	p.applied = true
	return result, nil
}

// Reconciliation can reuse openings that were already authenticated by explicit refresh.
func (s *WalletSession) storedWalletResolver(ctx context.Context) walletResolver {
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	return func(method wallet.Method, params wallet.DecryptParams) (any, error) {
		if err != nil {
			return nil, err
		}
		commitment := common.HexToHash(params.Context.OutputCommitment)
		var plaintext []byte
		if method == wallet.DecryptNote {
			if n, ok := state.Notes[commitment]; ok {
				plaintext = core.EncodeNotePlaintext(n.Note)
			}
		} else {
			for _, c := range state.PostpaidCredit.Grants {
				role := "owner"
				if c.Role == credit.Spender {
					role = "spender"
				}
				if c.Commitment == commitment && role == params.Context.Role {
					plaintext = credit.EncodeDeliveryPlaintext(c.Delivery)
					break
				}
			}
		}
		if plaintext == nil {
			return nil, wallet.ErrNotRecipient
		}
		c, err := params.Context.ProtocolContext.Core()
		if err != nil {
			return nil, err
		}
		token, err := wallet.Element(params.Context.TokenAddress)
		if err != nil {
			return nil, err
		}
		leaf, err := wallet.Element(common.HexToHash(params.Context.OutputCommitment).Big().String())
		if err != nil {
			return nil, err
		}
		if method == wallet.DecryptNote {
			return core.DecodeNotePlaintext(plaintext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: c, OutputCommitment: leaf}, TokenAddress: token, OwnerAddress: core.BigIntToElement(s.wallet.OwnerAddress().Big())})
		}
		role := credit.Owner
		if params.Context.Role == "spender" {
			role = credit.Spender
		}
		return credit.DecodeDeliveryPlaintext(plaintext, credit.DeliveryContext{ProtocolContext: c, TokenAddress: token, Commitment: leaf, RecipientIdentity: core.BigIntToElement(s.wallet.OwnerAddress().Big())}, role, s.wallet.ViewingPublicKey())
	}
}

func (p *PendingWalletRefresh) String() string   { return "PendingWalletRefresh([staged pages])" }
func (p *PendingWalletRefresh) GoString() string { return p.String() }
