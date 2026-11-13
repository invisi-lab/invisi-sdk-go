package broadcaster

import (
	"context"
	"math/big"
	"testing"
)

func TestQuoteCapabilityRejectsCopiesAndDetachesOptions(t *testing.T) {
	client := &Client{}
	var target QuoteTarget
	calls := 0
	BindQuoteTarget(&target, func(ctx context.Context, actual *Client, options QuoteOptions) error {
		calls++
		if actual != client {
			t.Fatal("client identity changed")
		}
		options.MaxFeeAmount.SetInt64(99)
		options.Schedule.ExecuteBefore = "99"
		return nil
	})
	options := QuoteOptions{MaxFeeAmount: big.NewInt(5), Schedule: &ScheduleOptions{PublishAfter: "10", ExecuteBefore: "20"}}
	copied := target
	if err := client.Quote(context.Background(), &copied, options); err == nil || calls != 0 {
		t.Fatal("copied capability invoked its original owner's callback", err)
	}
	if err := client.Quote(context.Background(), &target, options); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || options.MaxFeeAmount.Int64() != 5 || options.Schedule.ExecuteBefore != "20" {
		t.Fatal("quote changed caller-owned options")
	}
	var nilTarget *QuoteTarget
	if err := client.Quote(context.Background(), nilTarget, options); err == nil {
		t.Fatal("nil target accepted")
	}
	if err := client.Quote(context.Background(), &QuoteTarget{}, options); err == nil {
		t.Fatal("unbound target accepted")
	}
}
