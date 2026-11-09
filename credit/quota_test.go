package credit

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"os"
	"testing"
)

func TestQuotaDrawAndModeTransitions(t *testing.T) {
	s, _ := fixture(t)
	s.AvailableCredit = core.ElementFromUint64(7)
	s.Unlimited = false
	draw, err := s.Draw(core.ElementFromUint64(7), 2000000000)
	if err != nil || !draw.State.AvailableCredit.IsZero() || draw.State.TotalSpent != core.ElementFromUint64(7) {
		t.Fatal("gross draw must exhaust finite quota and preserve usage", err)
	}
	if _, err = draw.State.Draw(core.ElementFromUint64(1), 2000000001); err == nil {
		t.Fatal("finite zero authorized a draw")
	}
	on, off := true, false
	changed, err := draw.State.Update(Change{Policy: s.Policy, Expiry: s.Expiry, Unlimited: &on}, 2000000010)
	if err != nil || !changed.State.Unlimited || !changed.State.AvailableCredit.IsZero() {
		t.Fatal("unlimited must clear quota", err)
	}
	again, err := changed.State.Draw(core.ElementFromUint64(7), 2000000020)
	if err != nil || !again.State.AvailableCredit.IsZero() || again.State.TotalSpent != core.ElementFromUint64(14) {
		t.Fatal("unlimited draw must retain zero balance and add history", err)
	}
	exhausted, err := again.State.Update(Change{Policy: s.Policy, Expiry: s.Expiry, Unlimited: &off}, 2000000030)
	if err != nil || exhausted.State.Unlimited || !exhausted.State.AvailableCredit.IsZero() {
		t.Fatal("switch to finite must start exhausted without increase", err)
	}
	finite, err := again.State.Update(Change{Policy: s.Policy, Expiry: s.Expiry, Unlimited: &off, IncreaseAvailableCredit: core.ElementFromUint64(5)}, 2000000030)
	if err != nil || finite.State.Unlimited || finite.State.AvailableCredit != core.ElementFromUint64(5) || finite.State.TotalSpent != core.ElementFromUint64(14) {
		t.Fatal("finite mode must start at zero plus delta", err)
	}
}

func TestQuotaUpdateRejectsInvalidDeltasAndRetainsOriginalChange(t *testing.T) {
	s, _ := fixture(t)
	s.AvailableCredit = core.ElementFromUint64(10)
	on := true
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	for _, change := range []Change{
		{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: core.ElementFromUint64(1), DecreaseAvailableCredit: core.ElementFromUint64(1)},
		{Policy: s.Policy, Expiry: s.Expiry, DecreaseAvailableCredit: core.ElementFromUint64(11)},
		{Policy: s.Policy, Expiry: s.Expiry, Unlimited: &on, IncreaseAvailableCredit: core.ElementFromUint64(1)},
		{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: core.BigIntToElement(max)},
		{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: core.BigIntToElement(new(big.Int).Lsh(big.NewInt(1), 128))},
	} {
		if _, err := s.Update(change, 2000000010); err == nil {
			t.Fatal("accepted invalid quota delta")
		}
	}
	cleared, err := s.Update(Change{Policy: s.Policy, Expiry: s.Expiry, Unlimited: &on}, 2000000010)
	if err != nil || !cleared.State.Unlimited || !cleared.State.AvailableCredit.IsZero() {
		t.Fatal("switch to unlimited failed to clear available balance", err)
	}
	change := Change{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: core.ElementFromUint64(5)}
	a, err := s.Update(change, 2000000010)
	if err != nil {
		t.Fatal(err)
	}
	s.AvailableCredit = core.ElementFromUint64(3)
	b, err := s.Update(change, 2000000010)
	if err != nil {
		t.Fatal(err)
	}
	if a.State.AvailableCredit != core.ElementFromUint64(15) || b.State.AvailableCredit != core.ElementFromUint64(8) {
		t.Fatal("change restored stale quota instead of applying delta")
	}
	s.Unlimited = true
	if err = s.Validate(); err == nil {
		t.Fatal("unlimited state retained finite quota")
	}
}

func TestUpdateChangeSurvivesPersistenceAndRepricing(t *testing.T) {
	s, _ := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(core.ElementToBigInt(s.TokenAddress)), Timestamp: 2000000000}
	change := Change{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: core.ElementFromUint64(5)}
	transition, err := s.Update(change, 2000000300)
	if err != nil {
		t.Fatal(err)
	}
	clue := make([]byte, 68)
	clue[64] = 24
	deliveries := []protocol.Delivery{{Ciphertext: make([]byte, 485), Clue: clue}, {Ciphertext: make([]byte, 485), Clue: clue}}
	params := UpdateParams{Token: c.TokenAddress, Owner: common.BigToAddress(core.ElementToBigInt(s.OwnerWalletIdentity)), Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("1")}, Nullifier: common.Hash(transition.Nullifier.Bytes()), CreditCommitment: common.Hash(transition.StateCommitment.Bytes()), RevocationShard: 1, Deadline: big.NewInt(2000000300)}
	prepared, err := PrepareUpdate(c, params, deliveries, LocatedState{State: s, TreeID: 1}, change)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(intent)
	if err != nil {
		t.Fatal(err)
	}
	var restored Intent
	if err = json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	for _, amount := range []int64{1, 2, 3} {
		restored, err = reprice(restored, common.HexToAddress("0x1234"), big.NewInt(amount))
		if err != nil {
			t.Fatal(err)
		}
		reopened, err := Restore(context.Background(), restored, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		preview := reopened.Preview()
		if preview.AvailableCredit.Cmp(big.NewInt(405)) != 0 || preview.IncreaseAvailableCredit.Cmp(big.NewInt(5)) != 0 {
			t.Fatal("reprice applied delta repeatedly")
		}
	}
	restored.Inputs[0].State.AvailableCredit = core.ElementFromUint64(1)
	if _, err = Restore(context.Background(), restored, nil, nil); err == nil {
		t.Fatal("restored stale successor with substituted input balance")
	}
}

func TestCanonicalQuotaModeAndDeltaVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/quota.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Name, Kind string
			Input      map[string]any
			Changes    map[string]any
			Debit      string
			Result     struct {
				Nullifier, StateCommitment string
				State                      map[string]any
			}
		}
	}
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, v := range fixtures.Cases {
		t.Run(v.Name, func(t *testing.T) {
			input := vectorState(t, v.Input)
			var transition Transition
			var err error
			if v.Kind == "draw" {
				debit, e := canonicalField(v.Debit)
				if e != nil {
					t.Fatal(e)
				}
				transition, err = input.Draw(debit, 2000000000)
			} else {
				change := Change{Policy: input.Policy, Expiry: input.Expiry}
				if unlimited, ok := v.Changes["unlimited"]; ok {
					flag := unlimited.(bool)
					change.Unlimited = &flag
				}
				if increase, ok := v.Changes["increaseAvailableCredit"]; ok {
					change.IncreaseAvailableCredit, err = canonicalField(increase.(string))
					if err != nil {
						t.Fatal(err)
					}
				}
				if decrease, ok := v.Changes["decreaseAvailableCredit"]; ok {
					change.DecreaseAvailableCredit, err = canonicalField(decrease.(string))
					if err != nil {
						t.Fatal(err)
					}
				}
				transition, err = input.Update(change, 2000000300)
			}
			if err != nil {
				t.Fatal(err)
			}
			expected := vectorState(t, v.Result.State)
			if hex(transition.Nullifier) != v.Result.Nullifier || hex(transition.StateCommitment) != v.Result.StateCommitment || transition.State.AvailableCredit != expected.AvailableCredit || transition.State.Unlimited != expected.Unlimited || transition.State.Usage != expected.Usage {
				t.Fatal("canonical quota transition differs")
			}
		})
	}
}
