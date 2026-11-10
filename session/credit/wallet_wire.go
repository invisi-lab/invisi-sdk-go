package session

import (
	"context"

	wire "github.com/invisi-lab/invisi-sdk-go/wallet/protocol"
)

func (o *Operation) ExportWalletRequest(ctx context.Context, id string) (wire.Request, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return wire.Request{}, err
	}
	return o.prepared.ExportWalletRequest(id)
}
func (o *Operation) WithWalletResponse(ctx context.Context, q wire.Request, r wire.Response) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	return o.prepared.WithWalletResponse(q, r)
}
