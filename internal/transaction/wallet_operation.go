package transaction

import (
	"context"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func (o *Execution) ExportWalletRequest(ctx context.Context, id string) (wallet.Request, error) {
	request, err := o.WalletRequest(ctx)
	if err != nil {
		return wallet.Request{}, err
	}
	return o.Runtime.Wallet.PrepareRequest(id, wallet.Authorize, wallet.Authorization(request))
}
func (o *Execution) AttachWalletResponse(ctx context.Context, q wallet.Request, response wallet.Response) error {
	current, err := o.WalletRequest(ctx)
	if err != nil {
		return err
	}
	if q.Method != wallet.Authorize {
		return errors.New("wallet authorization response required")
	}
	value, err := o.Runtime.Wallet.AcceptResponse(q, response)
	if err != nil {
		return err
	}
	witness := value.(core.WalletAuthorizationWitness)
	if err = VerifyAuthorizationWitness(o.Runtime.Wallet.OwnerAddress(), current, witness); err != nil {
		return err
	}
	o.Authorization = &witness
	o.Status = "authorized"
	return o.Persist(ctx, o.Status)
}
func (p *PreparedTransaction) ExportWalletRequest(ctx context.Context, id string) (wallet.Request, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return wallet.Request{}, err
	}
	return p.operation.ExportWalletRequest(ctx, id)
}
func (p *PreparedTransaction) WithWalletResponse(ctx context.Context, q wallet.Request, r wallet.Response) (*PreparedTransaction, error) {
	q.Params = append([]byte(nil), q.Params...)
	r.Result = append([]byte(nil), r.Result...)
	return p.stage(ctx, func(o *Execution) error { return o.AttachWalletResponse(ctx, q, r) })
}
