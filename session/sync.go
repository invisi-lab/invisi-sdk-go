package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

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
		for id, record := range state.PostpaidCredit.Operations {
			if record.Externalized {
				record.Status = "reorged"
				record.WorkerID = ""
				state.PostpaidCredit.Operations[id] = record
			}
		}
		return nil
	})
}

type RefreshOptions struct {
	FromBlock uint64
	ToBlock   *uint64
	Reset     bool
	PageSize  uint64
}

type RefreshResult struct {
	Block                    *protocolrpc.CanonicalBlock
	AddedNotes, AddedCredits int
}

// Replay unresolved Credit reservations after a reset without releasing externalized authority.
func restoreCreditReservations(state *creditsession.State) error {
	for id, record := range state.Operations {
		if record.Status == "confirmed" || record.Status == "reverted" || record.Status == "cancelled" {
			continue
		}
		for _, input := range record.Intent.Inputs {
			c, e := input.State.Commitment()
			if e != nil {
				return e
			}
			commitment := common.Hash(c.Bytes())
			for key, grant := range state.Grants {
				if grant.Commitment != commitment || grant.Consumed {
					continue
				}
				if grant.ReservedBy != "" && grant.ReservedBy != id {
					return errors.New("conflicting persisted Credit reservations")
				}
				grant.ReservedBy = id
				state.Grants[key] = grant
			}
		}
	}
	return nil
}
