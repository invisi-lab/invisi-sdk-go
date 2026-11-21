package credit

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
)

func periodTime(t *testing.T, value string) uint64 {
	t.Helper()
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return uint64(date.Unix())
}

// Losing either period comparison, or treating a period rollover as expiry,
// must break both operations even when their business deadline remains valid.
func TestCreditPeriodRolloverRequiresReauthorization(t *testing.T) {
	for _, operation := range []string{"initial-payment", "draw"} {
		for _, boundary := range []struct{ name, prepared, submit string }{
			{"day", "2026-11-20T23:59:59Z", "2026-11-21T00:00:00Z"},
			{"month", "2026-11-30T23:59:59Z", "2026-12-01T00:00:00Z"},
			{"year", "2026-12-31T23:59:59Z", "2027-01-01T00:00:00Z"},
		} {
			t.Run(operation+"/"+boundary.name, func(t *testing.T) {
				p := periodPrepared(t, operation, periodTime(t, boundary.prepared), periodTime(t, boundary.submit)+3600)
				err := p.CheckTimestamp(periodTime(t, boundary.submit))
				if !errors.Is(err, ErrCreditPeriodChanged) {
					t.Fatalf("expected period error, got %v", err)
				}
				var changed *CreditPeriodChangedError
				if !errors.As(err, &changed) || changed.Code != "CREDIT_PERIOD_CHANGED" {
					t.Fatalf("missing stable code: %v", err)
				}
				if !strings.Contains(err.Error(), "reauthorize") || !strings.Contains(err.Error(), "regenerate") {
					t.Fatalf("missing recovery instruction: %v", err)
				}
			})
		}
		t.Run(operation+"/same-period-and-expiry", func(t *testing.T) {
			now := periodTime(t, "2026-11-21T12:00:00Z")
			p := periodPrepared(t, operation, now, now+60)
			if err := p.CheckTimestamp(now + 60); err != nil {
				t.Fatal(err)
			}
			err := p.CheckTimestamp(now + 61)
			if err == nil || errors.Is(err, ErrCreditPeriodChanged) || !strings.Contains(err.Error(), "expired") {
				t.Fatalf("expiry lost: %v", err)
			}
		})
	}
}

func periodPrepared(t *testing.T, operation string, timestamp, deadline uint64) *Prepared {
	t.Helper()
	b, state, _ := requestBuilder(t)
	b.Frame = func(_ context.Context, token common.Address, _ uint64) (Context, uint64, error) {
		return Context{ProtocolContext: state.ProtocolContext, TokenAddress: token, Timestamp: timestamp}, deadline, nil
	}
	token := common.BigToAddress(primitives.ElementToBigInt(state.TokenAddress))
	var p *Prepared
	var err error
	if operation == "initial-payment" {
		p, err = b.Create(context.Background(), CreatePreparation{TokenAddress: token, Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(90), InitialPayment: &InitialPayment{Amount: big.NewInt(7)}})
	} else {
		original := b.Input
		b.Input = func(ctx context.Context, c Context, h common.Hash, r Role) (PreparationInput, error) {
			input, e := original(ctx, c, h, r)
			input.Located.State.Usage = Usage{}
			input.Delivery.State.Usage = Usage{}
			return input, e
		}
		p, err = b.Draw(context.Background(), DrawPreparation{TokenAddress: token, Credits: []DrawMember{{Credit: common.HexToHash("0x1"), Amount: big.NewInt(7)}}})
	}
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDetachedCreditRejectsChangedPeriodBeforeTransport(t *testing.T) {
	for _, operation := range []string{"initial-payment", "draw"} {
		t.Run(operation, func(t *testing.T) {
			before := periodTime(t, "2026-11-20T23:59:59Z")
			p := periodPrepared(t, operation, before, before+3600)
			client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: "https://unused.example", QuoteSigner: common.HexToAddress("0x1"), FeeRecipient: common.HexToAddress("0x2")})
			if err != nil {
				t.Fatal(err)
			}
			p.intent.Broadcaster = &BroadcasterBinding{Configuration: client.Configuration()}
			reads := 0
			p.SetChainTimestampReader(func(context.Context) (uint64, error) { reads++; return before + 1, nil })
			_, err = client.Submit(context.Background(), p)
			if !errors.Is(err, ErrCreditPeriodChanged) || reads != 1 {
				t.Fatalf("fresh chain guard missing: reads=%d err=%v", reads, err)
			}
		})
	}
}

func TestDetachedCreditRequiresAvailableChainTimestamp(t *testing.T) {
	p := periodPrepared(t, "draw", 1, 300)
	if err := p.CheckCurrentTimestamp(context.Background()); err == nil {
		t.Fatal("missing chain reader accepted")
	}
	unavailable := errors.New("chain unavailable")
	p.SetChainTimestampReader(func(context.Context) (uint64, error) { return 0, unavailable })
	if err := p.CheckCurrentTimestamp(context.Background()); !errors.Is(err, unavailable) {
		t.Fatal(err)
	}
	p.SetChainTimestampReader(func(context.Context) (uint64, error) { return 0, nil })
	if err := p.CheckCurrentTimestamp(context.Background()); err == nil {
		t.Fatal("zero chain timestamp accepted")
	}
}

func TestCreditScheduleRejectsNextDayFrozenProof(t *testing.T) {
	for _, operation := range []string{"initial-payment", "draw"} {
		t.Run(operation, func(t *testing.T) {
			now := uint64(time.Now().Unix())
			tomorrow := (now/86400 + 1) * 86400
			p := periodPrepared(t, operation, now, tomorrow+3600)
			state, _ := fixture(t)
			p.intent.Context.ProtocolContext = state.ProtocolContext
			key, _ := crypto.GenerateKey()
			client := creditBroadcaster(t, p.intent.Context, key, common.HexToAddress("0x1234"))
			err := client.Quote(context.Background(), p, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2), Schedule: &broadcaster.ScheduleOptions{PublishAfter: strconv.FormatUint(tomorrow, 10)}})
			if !errors.Is(err, ErrCreditPeriodChanged) {
				t.Fatalf("next UTC day scheduled: %v", err)
			}
		})
	}
}

func TestSamePeriodScheduleRetainsBusinessDeadlineCoverageError(t *testing.T) {
	for _, operation := range []string{"initial-payment", "draw"} {
		t.Run(operation, func(t *testing.T) {
			// Prepare at the publication time to isolate same-period coverage even near midnight.
			after := (uint64(time.Now().Unix())/86400+1)*86400 + 12*3600
			p := periodPrepared(t, operation, after, after+10)
			key, _ := crypto.GenerateKey()
			client := creditBroadcaster(t, p.intent.Context, key, common.HexToAddress("0x1234"))
			err := client.Quote(context.Background(), p, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2), Schedule: &broadcaster.ScheduleOptions{PublishAfter: strconv.FormatUint(after, 10), ExecuteBefore: strconv.FormatUint(after+20, 10)}})
			if err == nil || !strings.Contains(err.Error(), "intent deadline does not cover scheduled window and slack") {
				t.Fatalf("lost business deadline error: %v", err)
			}
		})
	}
}
