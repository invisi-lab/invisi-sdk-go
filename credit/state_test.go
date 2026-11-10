package credit

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func fixture(t *testing.T) (State, map[string]any) {
	t.Helper()
	raw, err := os.ReadFile("testdata/state.json")
	if err != nil {
		t.Fatal(err)
	}
	var f map[string]any
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return vectorState(t, f["state"].(map[string]any)), f
}
func vectorState(t *testing.T, n map[string]any) State {
	t.Helper()
	e := func(k string) fr.Element {
		var v fr.Element
		if _, err := v.SetString(n[k].(string)); err != nil {
			t.Fatal(err)
		}
		return v
	}
	return State{AvailableCredit: e("availableCredit"), Unlimited: n["unlimited"].(bool), ProtocolContext: primitives.ProtocolContext{ChainID: primitives.ElementFromUint64(31337), ProtocolAddress: primitives.ElementFromUint64(17)}, TokenAddress: primitives.ElementFromUint64(34), SpenderWalletIdentity: e("spenderWalletIdentity"), OwnerWalletIdentity: e("ownerWalletIdentity"), RevocationShard: e("revocationShard"), RevocationSlot: e("revocationSlot"), Nonce: e("nonce"), Expiry: e("expiry"), SharedKey: e("creditSharedKey"), Policy: Policy{e("singleLimit"), e("dailyLimit"), e("monthlyLimit")}, Usage: Usage{e("lastDay"), e("dailySpent"), e("lastMonth"), e("monthlySpent"), e("totalSpent")}}
}

func hex(v fr.Element) string {
	b := v.Bytes()
	const chars = "0123456789abcdef"
	out := make([]byte, 66)
	copy(out, "0x")
	for i, x := range b {
		out[2+i*2] = chars[x>>4]
		out[3+i*2] = chars[x&15]
	}
	return string(out)
}
func TestCanonicalStateAndDraw(t *testing.T) {
	s, f := fixture(t)
	c, err := s.Commitment()
	if err != nil || hex(c) != f["commitment"] {
		t.Fatalf("commitment %s %v", hex(c), err)
	}
	next, err := s.Draw(primitives.ElementFromUint64(7), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	expected := f["draw"].(map[string]any)
	if hex(next.Nullifier) != expected["nullifier"] || hex(next.StateCommitment) != expected["stateCommitment"] || hex(next.RecoveryAmount) != expected["recoveryAmount"] {
		t.Fatalf("canonical transition mismatch: %+v", next)
	}
	recovered, err := s.Recover(next.Nullifier, next.StateCommitment, next.RecoveryAmount, 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Nonce != next.State.Nonce || recovered.TotalSpent != primitives.ElementFromUint64(7) || recovered.AvailableCredit != primitives.ElementFromUint64(393) {
		t.Fatal("recovery mismatch")
	}
	next.RecoveryAmount.Add(&next.RecoveryAmount, &next.RecoveryAmount)
	if _, err := s.Recover(next.Nullifier, next.StateCommitment, next.RecoveryAmount, 2000000000); err == nil {
		t.Fatal("accepted corrupt recovery")
	}
}
func TestQuotaAndPeriods(t *testing.T) {
	s, _ := fixture(t)
	s.AvailableCredit = primitives.ElementFromUint64(10)
	next, err := s.Draw(primitives.ElementFromUint64(7), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := next.State.Draw(primitives.ElementFromUint64(4), 2000000000+86400*40); err == nil {
		t.Fatal("quota restored across periods")
	}
	if _, err := next.State.Draw(primitives.ElementFromUint64(1), 1900000000); err == nil {
		t.Fatal("accepted period rollback")
	}
	s.AvailableCredit = fr.Element{}
	s.Unlimited = true
	s.TotalSpent = primitives.ElementFromUint64(1000)
	next, err = s.Draw(primitives.ElementFromUint64(7), 2000000000)
	if err != nil || next.State.TotalSpent != primitives.ElementFromUint64(1007) {
		t.Fatal("disabled limit lost usage", err)
	}
}

func TestUpdatePreservesUsageAndNamespace(t *testing.T) {
	s, _ := fixture(t)
	draw, err := s.Draw(primitives.ElementFromUint64(7), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	current := draw.State
	next, err := current.Update(Change{Policy: Policy{}, Expiry: primitives.ElementFromUint64(2200000000)}, 2000000100)
	if err != nil {
		t.Fatal(err)
	}
	if next.State.Usage != current.Usage || next.State.AvailableCredit != current.AvailableCredit || next.State.Unlimited != current.Unlimited || next.State.RevocationShard != current.RevocationShard || next.State.RevocationSlot != current.RevocationSlot || next.State.SharedKey != current.SharedKey || next.State.Nonce == current.Nonce {
		t.Fatal("update changed immutable identity or usage")
	}
	if _, err := current.Update(Change{Policy: Policy{}, Expiry: primitives.ElementFromUint64(1999999999)}, 2000000100); err == nil {
		t.Fatal("accepted expired successor")
	}
	if _, err := current.Update(Change{Policy: Policy{}, Expiry: primitives.ElementFromUint64(2200000000)}, 2100000001); err == nil {
		t.Fatal("accepted expired previous grant")
	}
}
