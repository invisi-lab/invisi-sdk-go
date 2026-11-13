package broadcaster

import (
	"context"
	"math/big"
	"testing"
)

func TestQuoteCapabilityRejectsCopiesAndDetachesOptions(t *testing.T) {
	client := &Client{}
	var target OperationTarget
	calls := 0
	BindOperationTarget(&target, func(ctx context.Context, actual *Client, options QuoteOptions) error {
		calls++
		if actual != client {
			t.Fatal("client identity changed")
		}
		options.MaxFeeAmount.SetInt64(99)
		options.Schedule.ExecuteBefore = "99"
		return nil
	}, nil, nil)
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
	var nilTarget *OperationTarget
	if err := client.Quote(context.Background(), nilTarget, options); err == nil {
		t.Fatal("nil target accepted")
	}
	if err := client.Quote(context.Background(), &OperationTarget{}, options); err == nil {
		t.Fatal("unbound target accepted")
	}
}

func TestOperationCapabilitySubmitCancelRejectsCopiesAndNil(t *testing.T) {
	client := &Client{}
	var target OperationTarget
	calls := 0
	action := func(context.Context, *Client) (Status, error) { calls++; return Status{State: "cancelled"}, nil }
	BindOperationTarget(&target, nil, action, action)
	copied := target
	var typedNil *OperationTarget
	for _, bad := range []*OperationTarget{&copied, typedNil, &OperationTarget{}} {
		if _, err := client.Submit(context.Background(), bad); err == nil {
			t.Fatal("invalid submit capability accepted")
		}
		if _, err := client.Cancel(context.Background(), bad); err == nil {
			t.Fatal("invalid cancel capability accepted")
		}
	}
	if calls != 0 {
		t.Fatal("invalid capability invoked owner")
	}
	if _, err := client.Submit(context.Background(), &target); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Cancel(context.Background(), &target); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatal(calls)
	}
}
