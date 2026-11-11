package session

import (
	"context"
	"errors"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func (s *WalletSession) RetryUnresolved(ctx context.Context, d wallet.DeliveryDecryptor) (int, error) {
	if d == nil {
		return 0, errors.New("App delivery decryptor required")
	}
	select {
	case s.refreshGate <- struct{}{}:
		defer func() { <-s.refreshGate }()
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	state, e := s.fencedSnapshot(ctx)
	if e != nil {
		return 0, e
	}
	if state.Invalidated {
		return 0, errors.New("wallet requires reset scan")
	}
	resolved := 0
	for id, u := range state.Scan.Unresolved {
		before, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			return resolved, e
		}
		next, e := CloneState(before)
		if e != nil {
			return resolved, e
		}
		base := s.deliveryResolver(ctx, d)
		stillUnresolved := false
		resolver := func(m wallet.Method, p wallet.DecryptParams) (any, error) {
			if walletFingerprint(m, p) != id {
				return nil, wallet.ErrNotRecipient
			}
			value, e := base(m, p)
			if errors.Is(e, wallet.ErrAuthenticationFailed) {
				stillUnresolved = true
				return nil, wallet.ErrNotRecipient
			}
			return value, e
		}
		ns, cs, _, e := s.readPublicTransaction(ctx, u.Transaction, u.Anchor, resolver)
		if e != nil {
			return resolved, e
		}
		if stillUnresolved {
			continue
		}
		for _, n := range ns {
			old := next.Notes[n.Commitment]
			n.ReservedBy = old.ReservedBy
			next.Notes[n.Commitment] = n
		}
		for _, c := range cs {
			k := creditsession.GrantKey(c.Commitment, c.Role)
			old := next.PostpaidCredit.Grants[k]
			c.ReservedBy = old.ReservedBy
			next.PostpaidCredit.Grants[k] = c
		}
		delete(next.Scan.Unresolved, id)
		if u.Method == wallet.DecryptCredit && len(cs) > 0 {
			target := u.Anchor.Number
			if next.Scan.IndexedHead != nil && next.Scan.IndexedHead.Number > target {
				target = next.Scan.IndexedHead.Number
			}
			if target > next.Scan.CreditReplayTarget {
				next.Scan.CreditReplayTarget = target
			}
			next.Scan.CreditReplayRevision++
			next.Scan.CreditReplayCursor = 0
		}
		if e = s.commitPrivateBatch(ctx, next.Scan.Generation, before.Revision, next); e != nil {
			return resolved, e
		}
		resolved++
	}
	return resolved, nil
}
