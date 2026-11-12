package credit

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// UpdatePatch uses nil for preservation; an explicitly empty payload clears application data.
type UpdatePatch struct {
	SingleLimit, DailyLimit, MonthlyLimit, Expiry    *fr.Element
	Unlimited                                        *bool
	IncreaseAvailableCredit, DecreaseAvailableCredit fr.Element
	Payload                                          []byte
}

func (p UpdatePatch) Change(s State) Change {
	c := Change{Policy: s.Policy, Expiry: s.Expiry, Unlimited: p.Unlimited, IncreaseAvailableCredit: p.IncreaseAvailableCredit, DecreaseAvailableCredit: p.DecreaseAvailableCredit, Payload: p.Payload}
	if p.SingleLimit != nil {
		c.Policy.SingleLimit = *p.SingleLimit
	}
	if p.DailyLimit != nil {
		c.Policy.DailyLimit = *p.DailyLimit
	}
	if p.MonthlyLimit != nil {
		c.Policy.MonthlyLimit = *p.MonthlyLimit
	}
	if p.Expiry != nil {
		c.Expiry = *p.Expiry
	}
	return c
}
