package session

import (
	"context"
	"errors"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"

	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

// Credit composes independent grant state and journals with this wallet's encrypted Store.
func (s *WalletSession) Credit() (*creditsession.Session, error) {
	if s == nil || s.options.Store == nil {
		return nil, errors.New("managed wallet store required")
	}
	context := s.coreContext()
	return creditsession.New(creditsession.Options{WrapPreparation: transaction.WrapPreparation, Build: s.buildCreditRequest, Store: creditStoreAdapter{session: s}, OwnerAddress: s.scope.OwnerAddress, Context: &context, Validate: s.validateCreditIntent, Prepare: s.prepareCreditIntent, ValidatePreparation: s.validateCreditPreparation, CheckPreparation: s.checkCreditPreparation, VerifyRevocation: s.verifyCreditRevocation})
}

type creditStoreAdapter struct{ session *WalletSession }

func (a creditStoreAdapter) ReadSnapshot(ctx context.Context) (creditsession.State, error) {
	state, err := a.session.fencedSnapshot(ctx)
	if err != nil {
		return creditsession.State{}, err
	}
	if state.Invalidated {
		return creditsession.State{}, errors.New("wallet requires reset scan")
	}
	if state.PostpaidCredit.Version == 0 {
		return creditsession.EmptyState(), nil
	}
	if state.PostpaidCredit.Version != 2 {
		return creditsession.State{}, errors.New("legacy Credit store is unsupported")
	}
	return creditsession.CloneState(state.PostpaidCredit)
}
func (a creditStoreAdapter) Update(ctx context.Context, change func(*creditsession.State) error) error {
	var generation *uint64
	if a.session.publicHistory != nil {
		public, err := a.session.publicHistory.Snapshot(ctx)
		if err != nil {
			return err
		}
		generation = &public.Generation
	}
	err := a.session.options.Store.Update(ctx, a.session.scope, func(state *WalletState) error {
		if state.Invalidated || generation != nil && state.Scan.Generation != *generation {
			return errors.New("wallet requires reset scan")
		}
		next := state.PostpaidCredit
		if next.Version == 0 {
			next = creditsession.EmptyState()
		}
		if next.Version != 2 {
			return errors.New("legacy Credit store is unsupported")
		}
		if next.Revision == ^uint64(0) {
			return errors.New("Credit revision exhausted")
		}
		if err := change(&next); err != nil {
			return err
		}
		next.Revision++
		detached, err := creditsession.CloneState(next)
		if err != nil {
			return err
		}
		state.PostpaidCredit = detached
		return nil
	})
	if err != nil {
		return err
	}
	if generation == nil {
		return nil
	}
	after, err := a.session.publicHistory.Snapshot(ctx)
	if err != nil {
		return err
	}
	if after.Generation != *generation {
		_ = a.session.invalidate(ctx)
		return errors.New("wallet requires reset scan")
	}
	return nil
}
