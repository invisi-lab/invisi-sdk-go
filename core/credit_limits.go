package core

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// CreditSpendingState is private commitment state. Zero limits disable their checks.
// Period zero is reserved for unused counters. Actual usage persists even when a limit is disabled.
type CreditSpendingState struct {
	SingleLimit, DailyLimit, MonthlyLimit        fr.Element
	LastDay, DailySpent, LastMonth, MonthlySpent fr.Element
}

// CreditPeriod identifies a UTC Gregorian calendar day and month, numbered from one.
type CreditPeriod struct{ CurrentDay, CurrentMonth uint64 }

// CreditPeriods derives periods from an explicit chain block timestamp. It does not use local time.
func CreditPeriods(timestamp uint64) CreditPeriod {
	days := timestamp / 86400
	z := days + 719468
	era := z / 146097
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	year := yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	var month uint64
	if mp < 10 {
		month = mp + 3
	} else {
		month = mp - 9
	}
	if month <= 2 {
		year++
	}
	return CreditPeriod{days + 1, (year-1970)*12 + month}
}
func (s CreditSpendingState) Fields() []fr.Element {
	return []fr.Element{s.SingleLimit, s.DailyLimit, s.MonthlyLimit, s.LastDay, s.DailySpent, s.LastMonth, s.MonthlySpent}
}
func ValidateCreditSpendingState(s CreditSpendingState) error {
	for _, v := range []fr.Element{s.SingleLimit, s.DailyLimit, s.MonthlyLimit} {
		if err := checkUint(v, 126, "Credit money"); err != nil {
			return err
		}
	}
	for _, v := range []fr.Element{s.DailySpent, s.MonthlySpent} {
		if err := checkUint(v, 128, "Credit usage"); err != nil {
			return err
		}
	}
	for _, v := range []fr.Element{s.LastDay, s.LastMonth} {
		if err := checkUint(v, 64, "Credit period"); err != nil {
			return err
		}
	}
	for _, p := range []struct{ period, spent fr.Element }{{s.LastDay, s.DailySpent}, {s.LastMonth, s.MonthlySpent}} {
		if p.period.IsZero() && !p.spent.IsZero() {
			return fmt.Errorf("unused Credit period requires zero spent")
		}
	}

	return nil
}

// ApplyCreditDrawUsage validates one input's complete debit, including full consumption,
// and returns a copy carrying updated counters. The amount/nonce are set by the caller.
func ApplyCreditDrawUsage(note CreditNote, amount fr.Element, timestamp uint64) (CreditNote, error) {
	if err := ValidateCreditSpendingState(note.CreditSpendingState); err != nil {
		return CreditNote{}, err
	}
	if err := checkUint(amount, 126, "draw amount"); err != nil {
		return CreditNote{}, err
	}
	if amount.IsZero() || amount.Cmp(&note.Amount) > 0 {
		return CreditNote{}, fmt.Errorf("draw must be positive and at most Credit balance")
	}
	if !note.SingleLimit.IsZero() && amount.Cmp(&note.SingleLimit) > 0 {
		return CreditNote{}, fmt.Errorf("Credit single limit exceeded")
	}
	periods := CreditPeriods(timestamp)
	child := note.Clone()
	update := func(limit, last, spent fr.Element, current uint64) (fr.Element, fr.Element, error) {
		previous := ElementToBigInt(last).Uint64()
		if previous > current {
			return fr.Element{}, fr.Element{}, fmt.Errorf("Credit period moved backwards")
		}
		total := new(big.Int).Set(ElementToBigInt(amount))
		if previous == current {
			total.Add(total, ElementToBigInt(spent))
		}
		if total.BitLen() > 128 {
			return fr.Element{}, fr.Element{}, fmt.Errorf("Credit usage exceeds uint128")
		}
		if !limit.IsZero() && total.Cmp(ElementToBigInt(limit)) > 0 {
			return fr.Element{}, fr.Element{}, fmt.Errorf("Credit period limit exceeded")
		}
		return ElementFromUint64(current), BigIntToElement(total), nil
	}
	var err error
	child.LastDay, child.DailySpent, err = update(note.DailyLimit, note.LastDay, note.DailySpent, periods.CurrentDay)
	if err != nil {
		return CreditNote{}, err
	}
	child.LastMonth, child.MonthlySpent, err = update(note.MonthlyLimit, note.LastMonth, note.MonthlySpent, periods.CurrentMonth)
	if err != nil {
		return CreditNote{}, err
	}
	return child, nil
}

// ComputeCreditNoteCommitment commits all financial, identity and private spending state.
func ComputeCreditNoteCommitment(note CreditNote) (fr.Element, error) {
	key, err := ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if err != nil {
		return fr.Element{}, err
	}
	return ComputeCreditCommitment(note.ChainID, note.SpenderWalletIdentity, note.OwnerWalletIdentity, note.SpenderPayout, note.OwnerPayout, note.TokenAddress, note.Amount, note.Nonce, note.Expiry, key, note.CreditSpendingState)
}
