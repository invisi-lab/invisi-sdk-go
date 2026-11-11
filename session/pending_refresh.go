package session

import (
	"context"
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

type walletResolver func(wallet.Method, wallet.DecryptParams) (any, error)

// PendingWalletRefresh retains canonical candidate pages until App decryption is validated.
type PendingWalletRefresh struct {
	Block   *protocolrpc.CanonicalBlock
	mu      sync.Mutex
	applied bool
	staged  *stagedPublicPage
	session *WalletSession
	base    WalletState
	options RefreshOptions
}

func walletFingerprint(m wallet.Method, p wallet.DecryptParams) string {
	return string(m) + ":" + string(wallet.Encode(p))
}
func deliveryRequest(envelope []byte, c primitives.ProtocolContext, token string, commitment common.Hash, role string) wallet.DecryptParams {
	return wallet.DecryptParams{Envelope: wallet.Hex(envelope), Context: wallet.DeliveryContext{ProtocolContext: wallet.Context(c), TokenAddress: token, OutputCommitment: commitment.Hex(), Role: role}}
}
func resolveNote(resolvers []walletResolver, envelope []byte, c primitives.NoteDecryptionContext) (primitives.Note, error) {
	if len(resolvers) == 0 || resolvers[0] == nil {
		return primitives.Note{}, errors.New("wallet response required")
	}
	p := deliveryRequest(envelope, c.ProtocolContext, c.TokenAddress.String(), common.Hash(c.OutputCommitment.Bytes()), "")
	p.Context.Owner = common.BigToAddress(primitives.ElementToBigInt(c.OwnerAddress)).Hex()
	p.Context.Clue = wallet.Hex(c.Clue)
	v, err := resolvers[0](wallet.DecryptNote, p)
	if err != nil {
		return primitives.Note{}, err
	}
	note, ok := v.(primitives.Note)
	if !ok {
		return primitives.Note{}, errors.New("invalid note response")
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
	p := deliveryRequest(envelope, c.ProtocolContext, c.TokenAddress.String(), common.Hash(c.Commitment.Bytes()), name)
	p.Context.Owner = common.BigToAddress(primitives.ElementToBigInt(c.RecipientIdentity)).Hex()
	p.Context.Clue = wallet.Hex(c.Clue)
	v, err := resolvers[0](wallet.DecryptCredit, p)
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
	return s.stageSharedRefresh(ctx, options)
}

func (p *PendingWalletRefresh) Decrypt(ctx context.Context, d wallet.DeliveryDecryptor) (RefreshResult, error) {
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
	current, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return RefreshResult{}, e
	}
	if current.Revision != p.base.Revision {
		return RefreshResult{}, errors.New("wallet changed during pending refresh")
	}
	result, e := s.scanHistory(ctx, d, ScanOptions{ToBlock: p.options.ToBlock, PageSize: p.options.PageSize, Reset: p.options.Reset, FromBlock: refreshLowerBound(p.options.FromBlock)}, p.staged)
	if e == nil {
		p.applied = true
	}
	return result, e
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
				plaintext = primitives.EncodeNotePlaintext(n.Note)
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
			return primitives.DecodeNotePlaintext(plaintext, primitives.NoteDecryptionContext{NoteEncryptionContext: primitives.NoteEncryptionContext{ProtocolContext: c, OutputCommitment: leaf}, TokenAddress: token, OwnerAddress: primitives.BigIntToElement(s.wallet.OwnerAddress().Big())})
		}
		role := credit.Owner
		if params.Context.Role == "spender" {
			role = credit.Spender
		}
		return credit.DecodeDeliveryPlaintext(plaintext, credit.DeliveryContext{ProtocolContext: c, TokenAddress: token, Commitment: leaf, RecipientIdentity: primitives.BigIntToElement(s.wallet.OwnerAddress().Big())}, role, s.wallet.ViewingPublicKey())
	}
}

func (p *PendingWalletRefresh) String() string   { return "PendingWalletRefresh([staged pages])" }
func (p *PendingWalletRefresh) GoString() string { return p.String() }
