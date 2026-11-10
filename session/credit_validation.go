package session

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

func (s *WalletSession) validateCreditIntent(ctx context.Context, intent credit.Intent) error {
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		return err
	}
	if state.Invalidated {
		return errors.New("wallet checkpoint invalidated; refresh before Credit authorization")
	}
	if err := s.protocol.ValidateDeployment(ctx, s.options.Provider); err != nil {
		return err
	}
	head, err := s.options.Provider.HeaderByNumber(ctx, nil)
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
	if err = prepared.CheckTimestamp(head.Time); err != nil {
		return err
	}
	contract, err := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, s.options.Provider)
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
		if p.FeeAmount != nil {
			fee.Set(p.FeeAmount)
		}
		next, err := contract.CreditAllocation(opts, p.Owner, p.Token)
		if err != nil {
			return err
		}
		if next.Shard != p.RevocationShard || next.Slot != p.RevocationSlot {
			return errors.New("Credit allocation changed; prepare again")
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
	for _, source := range sources {
		known, err := contract.RootHistory(opts, source.TreeId, source.MerkleRoot)
		if err != nil {
			return err
		}
		if !known {
			return errors.New("Credit source root is not registered")
		}
	}
	for _, input := range intent.Inputs {
		nf, err := input.State.Nullifier()
		if err != nil {
			return err
		}
		spent, err := contract.IsNullifierSpent(opts, nf.Bytes())
		if err != nil {
			return err
		}
		if spent {
			return errors.New("Credit input was consumed; refresh wallet")
		}
	}
	allocations := map[common.Address]*big.Int{}
	bitmaps := map[string]*big.Int{}
	checkGrant := func(owner common.Address, shard uint64, slot uint16) error {
		if shard == 0 {
			return errors.New("zero grant shard")
		}
		count := allocations[owner]
		if count == nil {
			next, err := contract.CreditAllocation(opts, owner, token)
			if err != nil {
				return err
			}
			if next.Shard == 0 {
				return errors.New("invalid current grant allocation")
			}
			count = new(big.Int).SetUint64(next.Shard - 1)
			count.Mul(count, big.NewInt(65536))
			count.Add(count, new(big.Int).SetUint64(uint64(next.Slot)))
			allocations[owner] = count
		}
		index := new(big.Int).SetUint64(shard - 1)
		index.Mul(index, big.NewInt(65536))
		index.Add(index, new(big.Int).SetUint64(uint64(slot)))
		if index.Cmp(count) >= 0 {
			return errors.New("grant reference is not allocated")
		}
		word := uint8(slot / 256)
		key := fmt.Sprintf("%s:%d:%d", owner.Hex(), shard, word)
		bits := bitmaps[key]
		if bits == nil {
			value, err := contract.CreditRevocationBitmap(opts, owner, token, shard, word)
			if err != nil {
				return err
			}
			bits = value
			bitmaps[key] = bits
		}
		if bits.Bit(int(slot%256)) != 0 {
			return errors.New("Credit authority was revoked")
		}
		return nil
	}
	if intent.Update != nil {
		p := intent.Update
		if err = checkGrant(p.Owner, p.RevocationShard, p.RevocationSlot); err != nil {
			return err
		}
	}
	if intent.Revoke != nil {
		p := intent.Revoke
		if err = checkGrant(p.Owner, p.Shard, p.Slot); err != nil {
			return err
		}
	}
	if intent.Draw != nil {
		for _, ref := range intent.Draw.GrantRefs {
			if err = checkGrant(intent.Draw.Payers[ref.PayerIndex.Uint64()].Payer, ref.Shard, ref.Slot); err != nil {
				return err
			}
		}
	}
	tokenCaller, err := contracts.NewIERC20Caller(token, s.options.Provider)
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
		for _, payer := range intent.Draw.Payers {
			if err = checkPayer(payer.Payer, payer.Amount); err != nil {
				return err
			}
		}
	} else if err = checkPayer(owner, fee); err != nil {
		return err
	}
	canonical, err := s.options.Provider.HeaderByNumber(ctx, head.Number)
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
	return operation.WithRevocationSignature(ctx, signature, verifier.VerifyTypedData)
}
