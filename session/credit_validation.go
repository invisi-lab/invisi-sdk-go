package session

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	transaction "github.com/invisi-labs/invisi-sdk-go/internal/transaction"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/prover"
	"github.com/invisi-labs/invisi-sdk-go/rpc/contracts"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

func (s *WalletSession) validateCreditIntent(ctx context.Context, intent credit.Intent) error {
	return s.validateCreditIntentWithProvider(ctx, intent, s.options.Provider, false)
}
func (s *WalletSession) prepareCreditIntent(ctx context.Context, intent credit.Intent) (*creditsession.PreparationEvidence, error) {
	snapshot, e := s.cachedPreparation(ctx)
	if e != nil {
		return nil, e
	}
	if snapshot != nil {
		e = s.validateCreditIntentWithProvider(ctx, intent, snapshot.Provider(s.options.Provider), true)
		if e == nil {
			return &creditsession.PreparationEvidence{Block: snapshot.Block, Generation: snapshot.Generation}, nil
		}
		if !errors.Is(e, transaction.ErrPreparationSnapshotMissing) {
			return nil, e
		}
	}
	return nil, s.validateCreditIntent(ctx, intent)
}
func (s *WalletSession) validateCreditIntentWithProvider(ctx context.Context, intent credit.Intent, provider transaction.Provider, historical bool) error {
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return err
	}
	if state.Invalidated {
		return errors.New("wallet checkpoint invalidated; refresh before Credit authorization")
	}
	if err := s.protocol.ValidateDeployment(ctx, provider); err != nil {
		return err
	}
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return errors.New("current Credit block unavailable")
	}
	prepared, err := credit.Restore(ctx, intent, nil, nil)
	if err != nil {
		return err
	}
	now := head.Time
	if historical && time.Now().Unix() > 0 && uint64(time.Now().Unix()) > now {
		now = uint64(time.Now().Unix())
	}
	if err = prepared.CheckTimestamp(now); err != nil {
		return err
	}
	contract, err := contracts.NewInvisiCoreCaller(s.scope.ProtocolAddress, provider)
	if err != nil {
		return err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
	token := intent.Context.TokenAddress
	var owner common.Address
	fee := new(big.Int)
	sources := []protocol.InputSource{}
	switch {
	case intent.Create != nil:
		p := intent.Create
		owner = p.Owner
		if p.HasInitialPayment != nil && *p.HasInitialPayment {
			fee.Set(p.TotalAmount)
		} else if p.FeeAmount != nil {
			fee.Set(p.FeeAmount)
		}
		status, err := contract.CreditStatusValue(opts, p.Owner, p.Token, p.GrantID)
		if err != nil {
			return err
		}
		if status != 0 {
			return errors.New("Credit grant ID already used or disabled")
		}
	case intent.Update != nil:
		p := intent.Update
		owner = p.Owner
		if p.FeeAmount != nil {
			fee.Set(p.FeeAmount)
		}
		sources = []protocol.InputSource{p.Source}
	case intent.Draw != nil:
		sources = intent.Draw.Sources
	case intent.Revoke != nil:
		p := intent.Revoke
		owner = p.Owner
		if p.FeeAmount != nil {
			fee.Set(p.FeeAmount)
		}
	default:
		return errors.New("Credit operation required")
	}
	_, err = transaction.MapPreparation(ctx, len(sources)+len(intent.Inputs), func(ctx context.Context, i int) (struct{}, error) {
		callOpts := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
		if i < len(sources) {
			source := sources[i]
			known, e := contract.RootHistory(callOpts, source.TreeId, source.MerkleRoot)
			if e != nil {
				return struct{}{}, e
			}
			if !known {
				return struct{}{}, errors.New("Credit source root is not registered")
			}
			return struct{}{}, nil
		}
		nf, e := intent.Inputs[i-len(sources)].State.Nullifier()
		if e != nil {
			return struct{}{}, e
		}
		spent, e := contract.IsNullifierSpent(callOpts, nf.Bytes())
		if e != nil {
			return struct{}{}, e
		}
		if spent {
			return struct{}{}, errors.New("Credit input was consumed; refresh wallet")
		}
		return struct{}{}, nil
	})
	if err != nil {
		return err
	}
	if err = checkCreditStatuses(ctx, contract, opts, token, intent); err != nil {
		return err
	}

	if historical {
		return nil
	}
	tokenCaller, err := contracts.NewIERC20Caller(token, provider)
	if err != nil {
		return err
	}
	checkPayer := func(payer common.Address, amount *big.Int) error {
		if amount.Sign() == 0 {
			return nil
		}
		balance, err := tokenCaller.BalanceOf(opts, payer)
		if err != nil {
			return err
		}
		allowance, err := tokenCaller.Allowance(opts, payer, s.scope.ProtocolAddress)
		if err != nil {
			return err
		}
		if balance.Cmp(amount) < 0 || allowance.Cmp(amount) < 0 {
			return errors.New("Credit payer requires sufficient ERC20 balance and Protocol allowance")
		}
		return nil
	}
	if intent.Draw != nil {
		_, err = transaction.MapPreparation(ctx, len(intent.Draw.Payers), func(_ context.Context, i int) (struct{}, error) {
			payer := intent.Draw.Payers[i]
			return struct{}{}, checkPayer(payer.Payer, payer.Amount)
		})
		if err != nil {
			return err
		}
	} else if err = checkPayer(owner, fee); err != nil {
		return err
	}
	canonical, err := provider.HeaderByNumber(ctx, head.Number)
	if err != nil {
		return err
	}
	if canonical == nil || canonical.Hash() != head.Hash() {
		return errors.New("Credit validation block changed")
	}
	return nil
}

func (s *WalletSession) ProveCredit(ctx context.Context, operation *creditsession.Operation, options prover.Options) error {
	if operation == nil {
		return errors.New("Credit operation required")
	}
	return operation.Prove(ctx, s.options.Prover, options)
}
func (s *WalletSession) WithCreditRevocationSignature(ctx context.Context, operation *creditsession.Operation, signature []byte) error {
	if operation == nil {
		return errors.New("Credit operation required")
	}
	verifier := transaction.GethFundingVerifier{Provider: s.options.Provider}
	return operation.WithRevocationSignature(ctx, signature, func(ctx context.Context, owner common.Address, typed apitypes.TypedData, sig []byte) (bool, error) {
		digest, _, e := apitypes.TypedDataAndHash(typed)
		if e != nil {
			return false, e
		}
		if recovered, ok := transaction.RecoverTypedDataSigner(digest, sig); ok && recovered == owner {
			return true, nil
		}
		return verifier.VerifyTypedData(ctx, owner, typed, sig)
	})
}

func (s *WalletSession) verifyCreditRevocation(ctx context.Context, owner common.Address, typed apitypes.TypedData, signature []byte) (bool, error) {
	provider := s.options.Provider
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return false, err
	}
	if head == nil || head.Number == nil {
		return false, errors.New("current Credit signature block unavailable")
	}
	valid, err := (transaction.GethFundingVerifier{Provider: provider}).VerifyTypedData(ctx, owner, typed, signature)
	if err != nil || !valid {
		return valid, err
	}
	canonical, err := provider.HeaderByNumber(ctx, head.Number)
	if err != nil {
		return false, err
	}
	if canonical == nil || canonical.Hash() != head.Hash() {
		return false, errors.New("Credit signature validation block changed")
	}
	return true, nil
}

func (s *WalletSession) checkCreditPreparation(ctx context.Context, evidence creditsession.PreparationEvidence) error {
	state, err := s.fencedSnapshot(ctx)
	if err != nil {
		return err
	}
	if state.Invalidated || state.Scan.Generation != evidence.Generation {
		_ = s.invalidate(ctx)
		return errors.New("Credit preparation history changed")
	}
	return nil
}
func (s *WalletSession) validateCreditPreparation(ctx context.Context, evidence creditsession.PreparationEvidence) error {
	if err := s.checkCreditPreparation(ctx, evidence); err != nil {
		return err
	}
	header, err := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(evidence.Block.Number))
	if err != nil {
		return err
	}
	if header == nil || header.Hash() != evidence.Block.Hash {
		return errors.New("Credit preparation anchor changed; prepare again")
	}
	return nil
}

// 取消释放只检查规范链上的消费、撤销与分配事实；过期的签名不影响取消。
func (s *WalletSession) checkCreditCancellation(ctx context.Context, intent credit.Intent) error {
	provider := s.options.Provider
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return err
	}
	if state.Invalidated {
		return errors.New("wallet checkpoint invalidated; refresh before Credit authorization")
	}
	if err := s.protocol.ValidateDeployment(ctx, provider); err != nil {
		return err
	}
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return errors.New("current Credit block unavailable")
	}
	contract, err := contracts.NewInvisiCoreCaller(s.scope.ProtocolAddress, provider)
	if err != nil {
		return err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
	token := intent.Context.TokenAddress
	switch {
	case intent.Create != nil:
		p := intent.Create
		status, err := contract.CreditStatusValue(opts, p.Owner, p.Token, p.GrantID)
		if err != nil {
			return err
		}
		if status != 0 {
			return errors.New("Credit grant ID already used or disabled")
		}
	case intent.Update != nil:
	case intent.Draw != nil:
	case intent.Revoke != nil:
	default:
		return errors.New("Credit operation required")
	}
	_, err = transaction.MapPreparation(ctx, len(intent.Inputs), func(ctx context.Context, i int) (struct{}, error) {
		callOpts := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
		nf, e := intent.Inputs[i].State.Nullifier()
		if e != nil {
			return struct{}{}, e
		}
		spent, e := contract.IsNullifierSpent(callOpts, nf.Bytes())
		if e != nil {
			return struct{}{}, e
		}
		if spent {
			return struct{}{}, errors.New("Credit input was consumed; refresh wallet")
		}
		return struct{}{}, nil
	})
	if err != nil {
		return err
	}
	if err = checkCreditStatuses(ctx, contract, opts, token, intent); err != nil {
		return err
	}

	canonical, err := provider.HeaderByNumber(ctx, head.Number)
	if err != nil {
		return err
	}
	if canonical == nil || canonical.Hash() != head.Hash() {
		return errors.New("Credit validation block changed")
	}
	return nil
}

// checkCreditStatuses binds each check to the canonical numeric block in opts.
func checkCreditStatuses(ctx context.Context, contract *contracts.InvisiCoreCaller, opts *bind.CallOpts, token common.Address, intent credit.Intent) error {
	type reference struct {
		owner  common.Address
		id     *big.Int
		revoke bool
	}
	refs := []reference{}
	if p := intent.Update; p != nil {
		refs = append(refs, reference{p.Owner, p.GrantID, false})
	}
	if p := intent.Revoke; p != nil {
		refs = append(refs, reference{p.Owner, p.GrantID, true})
	}
	if p := intent.Draw; p != nil {
		for _, ref := range p.GrantRefs {
			refs = append(refs, reference{p.Payers[ref.PayerIndex.Uint64()].Payer, ref.GrantID, false})
		}
	}
	statuses := map[string]uint8{}
	for _, ref := range refs {
		if !credit.ValidGrantID(ref.id) {
			return errors.New("invalid Credit grant ID")
		}
		key := fmt.Sprintf("%s:%s:%s", ref.owner, token, ref.id)
		status, ok := statuses[key]
		if !ok {
			var err error
			status, err = contract.CreditStatusValue(opts, ref.owner, token, ref.id)
			if err != nil {
				return err
			}
			if status > 2 {
				return errors.New("invalid Credit status")
			}
			statuses[key] = status
		}
		if ref.revoke {
			if status == 2 {
				return errors.New("Credit grant already disabled")
			}
		} else if status != 1 {
			return errors.New("Credit grant is not active")
		}
	}
	return nil
}
