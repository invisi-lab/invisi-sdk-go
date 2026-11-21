package paymentrouter

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestDraftBuilderPreservesActualHashValidation(t *testing.T) {
	ctx := context.Background()
	c, e := NewClient(deployment(), &canonicalChain{newChain(), bi(7)})
	if e != nil {
		t.Fatal(e)
	}
	core := coreCall(t)
	a, e := strictArgs(coreMethod, core.Data)
	if e != nil {
		t.Fatal(e)
	}
	a[0] = []byte{}
	core.Data, e = packCall(coreMethod, a...)
	if e != nil {
		t.Fatal(e)
	}
	o := PaymentOptions{Core: core, Mode: RelayedToken, Payer: addr(9), TokenIn: addr(2), RefundRecipient: addr(9), Relayer: addr(10), AmountInMaximum: bi(1000), Deadline: bi(1900), Route: Route{Channel: ChannelNone}}
	draft, e := c.BuildPaymentDraft(ctx, o)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.BuildPayment(ctx, o); e == nil {
		t.Fatal("final builder accepted missing proof")
	}
	p, e := DecodeRouterCall(draft.Data)
	if e != nil {
		t.Fatal(e)
	}
	if e = ValidateDraftIntent(p, deployment(), bi(1000), bi(7)); e != nil {
		t.Fatal(e)
	}
	p.Intent.CoreRequestHash = common.Hash{}
	if e = ValidateDraftIntent(p, deployment(), bi(1000), bi(7)); e == nil {
		t.Fatal("draft omitted actual Core hash validation")
	}
}
