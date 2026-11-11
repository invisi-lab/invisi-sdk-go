package transaction

import (
	"context"
)

func (s *Runtime) PersistPlan(ctx context.Context, plan Plan) (*Execution, error) {
	if e := s.Canonical(ctx, plan.Block); e != nil {
		return nil, e
	}
	id, e := RandomOperationID()
	if e != nil {
		return nil, e
	}
	plan.ID = id
	plan.Version = 1
	plan.OwnerAddress = s.Wallet.ownerAddress
	plan.Namespace = s.Scope.Namespace
	plan, e = ClonePlan(plan)
	if e != nil {
		return nil, e
	}
	execution := &Execution{Runtime: s, Plan: plan, Status: "prepared", WorkerID: id}
	if s.OnComplete != nil {
		if e = s.OnComplete(ctx, execution); e != nil {
			return nil, e
		}
	}
	return execution, nil
}
