package session

import (
	"context"
)

func (o *Operation) WithWalletSignature(ctx context.Context, signature []byte) (*Operation, error) {
	copied := append([]byte(nil), signature...)
	return o.stage(ctx, func(next *Operation) error { return next.attachWallet(ctx, copied) })
}

func (o *Operation) WithFundingSignature(ctx context.Context, signature []byte) (*Operation, error) {
	copied := append([]byte(nil), signature...)
	return o.stage(ctx, func(next *Operation) error { return next.attachFunding(ctx, copied) })
}
