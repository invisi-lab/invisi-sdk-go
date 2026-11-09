package session

import (
	"context"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func (o *Operation) ExportWalletRequest(ctx context.Context, id string) (wallet.Request, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.execution().ExportWalletRequest(ctx, id)
}
func (o *Operation) WithWalletResponse(ctx context.Context, q wallet.Request, r wallet.Response) (*Operation, error) {
	q.Params = append([]byte(nil), q.Params...)
	r.Result = append([]byte(nil), r.Result...)
	return o.stage(ctx, func(next *Operation) error {
		execution := next.execution()
		err := execution.AttachWalletResponse(ctx, q, r)
		next.apply(execution)
		return err
	})
}
