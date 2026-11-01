package core

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func limitedCredit() CreditNote {
	return CreditNote{ProtocolContext: ProtocolContext{ElementFromUint64(1), ElementFromUint64(2)}, SpenderWalletIdentity: ElementFromUint64(3), OwnerWalletIdentity: ElementFromUint64(4), SpenderPayout: ElementFromUint64(3), OwnerPayout: ElementFromUint64(4), TokenAddress: ElementFromUint64(5), Amount: ElementFromUint64(100), Nonce: ElementFromUint64(6), Expiry: ElementFromUint64(2100000000), CreditSharedKey: ElementFromUint64(7), CreditSpendingState: CreditSpendingState{SingleLimit: ElementFromUint64(20), DailyLimit: ElementFromUint64(40), MonthlyLimit: ElementFromUint64(60)}}
}
func TestCreditPeriodsGregorianBoundaries(t *testing.T) {
	for _, v := range []struct{ timestamp, day, month uint64 }{{0, 1, 1}, {86399, 1, 1}, {86400, 2, 1}, {5097599, 59, 2}, {5097600, 60, 3}, {68169599, 789, 26}, {68169600, 790, 26}, {68256000, 791, 27}, {10367999, 120, 4}, {10368000, 121, 5}, {31535999, 365, 12}, {31536000, 366, 13}, {4107542399, 47541, 1562}, {4107542400, 47542, 1563}} {
		p := CreditPeriods(v.timestamp)
		if p.CurrentDay != v.day || p.CurrentMonth != v.month {
			t.Fatalf("timestamp%d: %+v", v.timestamp, p)
		}
	}
}
func TestCreditUsageLimitsAndImmutableParent(t *testing.T) {
	n := limitedCredit()
	p := CreditPeriods(2000000000)
	child, err := ApplyCreditDrawUsage(n, ElementFromUint64(10), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	if child.LastDay != ElementFromUint64(p.CurrentDay) || child.DailySpent != ElementFromUint64(10) || child.MonthlySpent != ElementFromUint64(10) || !n.DailySpent.IsZero() {
		t.Fatal("wrong usage or parent mutated")
	}
	child, err = ApplyCreditDrawUsage(child, ElementFromUint64(15), 2000000000)
	if err != nil || child.DailySpent != ElementFromUint64(25) {
		t.Fatal("same period accumulation", err)
	}
	if _, err = ApplyCreditDrawUsage(child, ElementFromUint64(21), 2000000000); err == nil {
		t.Fatal("single limit bypass")
	}
	if _, err = ApplyCreditDrawUsage(child, ElementFromUint64(16), 2000000000); err == nil {
		t.Fatal("daily limit bypass")
	}
	full := limitedCredit()
	full.Amount = ElementFromUint64(21)
	if _, err = ApplyCreditDrawUsage(full, full.Amount, 2000000000); err == nil {
		t.Fatal("full consumption bypass")
	}
	child.SingleLimit = fr.Element{}
	child.DailyLimit = ElementFromUint64(100)
	child.MonthlySpent = ElementFromUint64(59)
	if _, err = ApplyCreditDrawUsage(child, ElementFromUint64(2), 2000000000); err == nil {
		t.Fatal("monthly limit bypass")
	}
	if _, err = ApplyCreditDrawUsage(child, ElementFromUint64(1), 1999900000); err == nil {
		t.Fatal("period moved backwards")
	}
	reset, err := ApplyCreditDrawUsage(child, ElementFromUint64(2), 2003000000)
	if err != nil || reset.DailySpent != ElementFromUint64(2) || reset.MonthlySpent != ElementFromUint64(2) {
		t.Fatal("calendar reset", err)
	}
}
func TestCreditPolicyCanonicalValidation(t *testing.T) {
	n := limitedCredit()
	n.DailyLimit = fr.Element{}
	n.LastDay = ElementFromUint64(1)
	if ValidateCreditSpendingState(n.CreditSpendingState) != nil {
		t.Fatal("disabled state with preserved period rejected")
	}
	n = limitedCredit()
	n.DailySpent = ElementFromUint64(1)
	if ValidateCreditSpendingState(n.CreditSpendingState) == nil {
		t.Fatal("zero sentinel spent accepted")
	}
	n = limitedCredit()
	n.LastDay = ElementFromUint64(1)
	n.DailySpent = ElementFromUint64(41)
	if ValidateCreditSpendingState(n.CreditSpendingState) != nil {
		t.Fatal("lowered limit below usage must remain a valid frozen state")
	}
}

func TestCreditCommitmentBindsAllSevenSpendingFields(t *testing.T) {
	n := limitedCredit()
	n.LastDay = ElementFromUint64(1)
	n.DailySpent = ElementFromUint64(10)
	n.LastMonth = ElementFromUint64(1)
	n.MonthlySpent = ElementFromUint64(15)
	original, err := ComputeCreditNoteCommitment(n)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 7; i++ {
		changed := n
		fields := []*fr.Element{&changed.SingleLimit, &changed.DailyLimit, &changed.MonthlyLimit, &changed.LastDay, &changed.DailySpent, &changed.LastMonth, &changed.MonthlySpent}
		one := ElementFromUint64(1)
		fields[i].Add(fields[i], &one)
		leaf, err := ComputeCreditNoteCommitment(changed)
		if err != nil || leaf == original {
			t.Fatalf("unbound field%d: %v", i, err)
		}
	}
	zero := n
	zero.CreditSpendingState = CreditSpendingState{}
	key, _ := ComputeCreditSharedKeyHash(zero.CreditSharedKey)
	old, _ := Poseidon2Hash(ElementFromUint64(DCredit), zero.ChainID, zero.SpenderWalletIdentity, zero.OwnerWalletIdentity, zero.SpenderPayout, zero.OwnerPayout, zero.TokenAddress, zero.Amount, zero.Nonce, zero.Expiry, key)
	current, err := ComputeCreditNoteCommitment(zero)
	if err != nil || current == old {
		t.Fatal("legacy11-field commitment accepted", err)
	}
}
func TestCreditRecoveryRequiresConfirmedPeriodAndPreservesUsage(t *testing.T) {
	n := limitedCredit()
	p := CreditPeriods(2000000000)
	n.LastDay = ElementFromUint64(p.CurrentDay)
	n.DailySpent = ElementFromUint64(10)
	n.LastMonth = ElementFromUint64(p.CurrentMonth)
	n.MonthlySpent = ElementFromUint64(15)
	leaf, err := ComputeCreditNoteCommitment(n)
	if err != nil {
		t.Fatal(err)
	}
	nf, _ := ComputeCreditNullifier(n.ChainID, n.CreditSharedKey, leaf)
	child, err := ApplyCreditDrawUsage(n, ElementFromUint64(15), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	child.Amount = ElementFromUint64(85)
	child.Nonce, _ = DeriveRemainderNonce(n.ChainID, n.ProtocolAddress, n.TokenAddress, nf, n.CreditSharedKey)
	cm, _ := ComputeCreditNoteCommitment(child)
	pad, _ := ComputeCreditRecoveryPad(n.ChainID, n.ProtocolAddress, n.TokenAddress, nf, n.CreditSharedKey)
	amount := child.Amount
	amount.Add(&amount, &pad)
	event := CreditRemainderRecovery{nf, cm, amount}
	recovered, err := RecoverCreditRemainder(n.ProtocolContext, n, event, 2000000000)
	if err != nil || recovered.CreditSpendingState != child.CreditSpendingState {
		t.Fatal("counter recovery", err)
	}
	if _, err = RecoverCreditRemainder(n.ProtocolContext, n, event, 2000086400); err == nil {
		t.Fatal("wrong confirmed period accepted")
	}
	n.Amount = ElementFromUint64(21)
	leaf, _ = ComputeCreditNoteCommitment(n)
	nf, _ = ComputeCreditNullifier(n.ChainID, n.CreditSharedKey, leaf)
	if _, err = RecoverCreditRemainder(n.ProtocolContext, n, CreditRemainderRecovery{CreditNullifier: nf}, 2000000000); err == nil {
		t.Fatal("full draw limit bypass")
	}
}
func TestCreditStateRangeAndCheckedSums(t *testing.T) {
	n := limitedCredit()
	overflow := new(big.Int).Lsh(big.NewInt(1), 126)
	n.SingleLimit = BigIntToElement(overflow)
	if ValidateCreditSpendingState(n.CreditSpendingState) == nil {
		t.Fatal("uint126 limit overflow")
	}
	n = limitedCredit()
	n.LastDay = BigIntToElement(new(big.Int).Lsh(big.NewInt(1), 64))
	if ValidateCreditSpendingState(n.CreditSpendingState) == nil {
		t.Fatal("uint64 period overflow")
	}
	maximum := new(big.Int).Sub(overflow, big.NewInt(1))
	n = limitedCredit()
	n.SingleLimit = fr.Element{}
	n.DailyLimit = BigIntToElement(maximum)
	n.DailySpent = n.DailyLimit
	n.LastDay = ElementFromUint64(CreditPeriods(2000000000).CurrentDay)
	if _, err := ApplyCreditDrawUsage(n, ElementFromUint64(1), 2000000000); err == nil {
		t.Fatal("counter overflow accepted")
	}
	p := CreditPeriods(^uint64(0))
	if p.CurrentDay != ^uint64(0)/86400+1 || p.CurrentMonth == 0 {
		t.Fatal("maximum timestamp overflow")
	}
}

func TestUnlimitedUsageSurvivesDisableAndReenable(t *testing.T) {
	n := limitedCredit()
	n.CreditSpendingState = CreditSpendingState{}
	first, err := ApplyCreditDrawUsage(n, ElementFromUint64(10), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	if first.DailySpent != ElementFromUint64(10) || first.MonthlySpent != ElementFromUint64(10) || first.LastDay.IsZero() || first.LastMonth.IsZero() {
		t.Fatal("unlimited debit was not recorded")
	}
	first.DailyLimit = ElementFromUint64(5)
	first.MonthlyLimit = fr.Element{}
	if ValidateCreditSpendingState(first.CreditSpendingState) != nil {
		t.Fatal("lowered policy must accept existing usage")
	}
	if _, err = ApplyCreditDrawUsage(first, ElementFromUint64(1), 2000000000); err == nil {
		t.Fatal("lowered daily limit did not freeze")
	}
	next, err := ApplyCreditDrawUsage(first, ElementFromUint64(1), 2000086400)
	if err != nil || next.DailySpent != ElementFromUint64(1) || next.MonthlySpent != ElementFromUint64(11) {
		t.Fatal("day rollover did not start fresh without resetting month", err)
	}
	first.DailyLimit = fr.Element{}
	disabled, err := ApplyCreditDrawUsage(first, ElementFromUint64(10), 2000000000)
	if err != nil || disabled.DailySpent != ElementFromUint64(20) {
		t.Fatal("disable reset history", err)
	}
	disabled.DailyLimit = ElementFromUint64(15)
	if _, err = ApplyCreditDrawUsage(disabled, ElementFromUint64(1), 2000000000); err == nil {
		t.Fatal("re-enabling forgot usage")
	}
	disabled.DailyLimit = fr.Element{}
	disabled.MonthlyLimit = ElementFromUint64(15)
	if _, err = ApplyCreditDrawUsage(disabled, ElementFromUint64(1), 2000086400); err == nil {
		t.Fatal("month usage reset at day rollover")
	}
	next, err = ApplyCreditDrawUsage(disabled, ElementFromUint64(1), 2003000000)
	if err != nil || next.MonthlySpent != ElementFromUint64(1) {
		t.Fatal("month rollover failed", err)
	}
}
func TestCreditUsageUsesCheckedUint128EvenWhenUnlimited(t *testing.T) {
	n := limitedCredit()
	n.CreditSpendingState = CreditSpendingState{LastDay: ElementFromUint64(CreditPeriods(2000000000).CurrentDay), LastMonth: ElementFromUint64(CreditPeriods(2000000000).CurrentMonth)}
	n.DailySpent = BigIntToElement(new(big.Int).Lsh(big.NewInt(1), 126))
	n.MonthlySpent = n.DailySpent
	next, err := ApplyCreditDrawUsage(n, ElementFromUint64(1), 2000000000)
	if err != nil || next.DailySpent == n.DailySpent {
		t.Fatal("valid uint128 usage rejected", err)
	}
	n.DailySpent = BigIntToElement(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	if _, err = ApplyCreditDrawUsage(n, ElementFromUint64(1), 2000000000); err == nil {
		t.Fatal("uint128 usage wrapped")
	}
	n.DailySpent = BigIntToElement(new(big.Int).Lsh(big.NewInt(1), 128))
	if ValidateCreditSpendingState(n.CreditSpendingState) == nil {
		t.Fatal("uint128 overflow accepted")
	}
}

func TestCreditV5PayloadCarriesUint128UsageAboveLimits(t *testing.T) {
	note, private, public, context := memoNote(t)
	wallet, err := DeriveWallet(private)
	if err != nil {
		t.Fatal(err)
	}
	credit := limitedCredit()
	credit.ProtocolContext = note.ProtocolContext
	credit.TokenAddress = note.TokenAddress
	credit.SpenderWalletIdentity = note.WalletIdentity
	credit.OwnerWalletIdentity = note.WalletIdentity
	credit.SpenderPayout = note.WalletIdentity
	credit.OwnerPayout = note.WalletIdentity
	credit.SingleLimit = ElementFromUint64(1)
	credit.DailyLimit = ElementFromUint64(1)
	credit.MonthlyLimit = fr.Element{}
	credit.LastDay = ElementFromUint64(1)
	credit.LastMonth = ElementFromUint64(1)
	credit.DailySpent = BigIntToElement(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(3)))
	credit.MonthlySpent = BigIntToElement(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(5)))
	cm, err := ComputeCreditNoteCommitment(credit)
	if err != nil {
		t.Fatal(err)
	}
	context.OutputCommitment = cm
	keys := ReceivingKeys{ViewingPublicKey: public, ClueKey: wallet.ClueKey}
	state := CreditDeliveryState{Note: credit, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}
	ciphertext, err := EncryptCreditNote(public, state, CreditNoteEncryptionContext{NoteEncryptionContext: context.NoteEncryptionContext, Role: CreditOwner})
	if err != nil {
		t.Fatal(err)
	}
	if len(ciphertext) != 498 || ciphertext[0] != 0x81 {
		t.Fatal("v5 wire changed")
	}
	decoded, err := DecryptCreditNote(private, ciphertext, CreditNoteDecryptionContext{NoteDecryptionContext: context, Role: CreditOwner})
	if err != nil || decoded.Note.CreditSpendingState != credit.CreditSpendingState {
		t.Fatal("uint128 usage lost in payload", err)
	}
}
