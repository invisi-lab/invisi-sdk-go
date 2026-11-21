package credit

import (
	"context"
	"fmt"
)

// CreditPeriodChangedError means the frozen usage statement no longer matches
// the chain's UTC day or month. A new authorization and proof are required.
type CreditPeriodChangedError struct{ Code string }

func (*CreditPeriodChangedError) Error() string {
	return "CREDIT_PERIOD_CHANGED: Credit usage period changed; reauthorize and regenerate the proof"
}
func (*CreditPeriodChangedError) Is(target error) bool {
	_, ok := target.(*CreditPeriodChangedError)
	return ok
}

// ErrCreditPeriodChanged supports errors.Is and carries the stable machine code.
var ErrCreditPeriodChanged = &CreditPeriodChangedError{Code: "CREDIT_PERIOD_CHANGED"}

// SetChainTimestampReader supplies the latest canonical chain block timestamp.
// Low-level and restored preparations need this reader before their first
// broadcaster submission. The reader is local capability and is never serialized.
func (p *Prepared) SetChainTimestampReader(reader func(context.Context) (uint64, error)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.chainTimestamp = reader
}

// CheckCurrentTimestamp fails closed for period-bound payments when chain time
// cannot be read. Pure create, update and revoke have no frozen usage period.
func (p *Prepared) CheckCurrentTimestamp(ctx context.Context) error {
	p.mu.Lock()
	bound := p.intent.Draw != nil || p.intent.Create != nil && p.intent.Create.HasInitialPayment != nil && *p.intent.Create.HasInitialPayment
	reader := p.chainTimestamp
	p.mu.Unlock()
	if !bound {
		return nil
	}
	if reader == nil {
		return fmt.Errorf("current Credit chain timestamp reader required")
	}
	timestamp, err := reader(ctx)
	if err != nil {
		return fmt.Errorf("read current Credit chain timestamp: %w", err)
	}
	if timestamp == 0 {
		return fmt.Errorf("current Credit chain timestamp unavailable")
	}
	return p.CheckTimestamp(timestamp)
}
