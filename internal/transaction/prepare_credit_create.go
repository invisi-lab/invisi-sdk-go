package transaction

import (
	"context"
)

func (s *Runtime) PrepareCreateCredit(ctx context.Context, p CreateCreditPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareCreateCredit", failure) }()
	var err error
	p, err = Frozen(p)
	if err != nil {
		return nil, err
	}
	return s.PrepareCreditCreation(ctx, p, nil, nil)
}
