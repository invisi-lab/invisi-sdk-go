package session

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	native "github.com/invisi-lab/invisi-sdk-go/internal/testsupport"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

var testMaterials sync.Map
var testExecutors sync.Map

func localTestWallet(key string) (*sdk.Wallet, error) {
	e, err := native.NewExecutor(key)
	if err != nil {
		return nil, err
	}
	q, err := wallet.Prepare("public-profile", e.OwnerAddress(), wallet.Public, struct{}{})
	if err != nil {
		return nil, err
	}
	r, err := e.ExecuteOne(context.Background(), q)
	if err != nil {
		return nil, err
	}
	w, err := sdk.WalletFromResponse(q, r)
	if err != nil {
		return nil, err
	}
	raw, err := primitives.DecodePrivateKey(key)
	if err != nil {
		return nil, err
	}
	defer clear(raw)
	m, err := primitives.DeriveWallet(raw)
	testMaterials.Store(w.OwnerAddress(), m)
	testExecutors.Store(w.OwnerAddress(), e)
	return w, err
}
func localTestMaterial(w *sdk.Wallet) *primitives.InvisiWallet {
	v, _ := testMaterials.Load(w.OwnerAddress())
	return v.(*primitives.InvisiWallet)
}

func testDisclosureViewer(s *WalletSession) *sdk.ViewingCapability {
	m := localTestMaterial(s.wallet)
	v, _ := sdk.NewViewingCapability(s.wallet.OwnerAddress(), m.ViewingPrivateKey, m.ViewingPublicKey)
	return v
}

func testExecutor(w *sdk.Wallet) *native.Executor {
	if w == nil {
		return nil
	}
	v, _ := testExecutors.Load(w.OwnerAddress())
	if v == nil {
		return nil
	}
	return v.(*native.Executor)
}

func refreshForTest(s *WalletSession, ctx context.Context, options RefreshOptions) (RefreshResult, error) {
	pending, err := s.Refresh(ctx, options)
	if err != nil {
		return RefreshResult{}, err
	}
	return pending.Decrypt(ctx, testDeliveryDecryptor(s))
}
func authorizeForTest(o *Operation, ctx context.Context) (*Operation, error) {
	if !o.plan.NeedsAuthorization {
		return o, nil
	}
	q, err := o.ExportWalletRequest(ctx, "test-authorize")
	if err != nil {
		return nil, err
	}
	r, err := testExecutor(o.session.wallet).ExecuteOne(ctx, q)
	if err != nil {
		return nil, err
	}
	return o.WithWalletResponse(ctx, q, r)
}
func readTransactionForTest(s *WalletSession, ctx context.Context, hash common.Hash, anchor protocolrpc.CanonicalBlock) ([]StoredNote, []creditsession.StoredState, []*authenticatedCreditDraw, error) {
	resolve := func(method wallet.Method, p wallet.DecryptParams) (any, error) {
		q, e := s.wallet.PrepareRequest("test-read", method, p)
		if e != nil {
			return nil, wallet.ErrNotRecipient
		}
		r, e := testExecutor(s.wallet).ExecuteOne(ctx, q)
		if e != nil {
			return nil, wallet.ErrNotRecipient
		}
		v, e := s.wallet.AcceptResponse(q, r)
		return v, e
	}
	return s.readTransaction(ctx, hash, anchor, resolve)
}

func testDeliveryDecryptor(s *WalletSession) wallet.DeliveryDecryptorFuncs {
	m := localTestMaterial(s.wallet)
	return wallet.DeliveryDecryptorFuncs{
		Note: func(ctx context.Context, b []byte, c primitives.NoteDecryptionContext) ([]byte, error) {
			n, err := primitives.DecryptPrivateTransferNote(m.ViewingPrivateKey, b, c)
			if err != nil {
				return nil, nil
			}
			return primitives.EncodeNotePlaintext(n), nil
		},
		Credit: func(ctx context.Context, b []byte, c credit.DeliveryContext, r credit.Role) ([]byte, error) {
			n, err := credit.Decrypt(m.ViewingPrivateKey, b, c, r)
			if err != nil {
				return nil, nil
			}
			return credit.EncodeDeliveryPlaintext(n), nil
		},
	}
}
