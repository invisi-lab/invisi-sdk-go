package sdk

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestDisclosuresInteroperateWithCurrentTSExports(t *testing.T) {
	raw, e := os.ReadFile("testdata/disclosure-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture map[string]json.RawMessage
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	wallet, e := WalletFromPrivateKey("0x" + strings.Repeat("0", 63) + "1")
	if e != nil {
		t.Fatal(e)
	}
	viewer, e := wallet.GetViewer()
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"noteOpening", "creditOpening", "notePayload", "creditPayload"} {
		t.Run(name, func(t *testing.T) {
			disclosure, e := DisclosureFromExport(fixture[name])
			if e != nil {
				t.Fatal(e)
			}
			output, e := disclosure.output()
			if e != nil {
				t.Fatal(e)
			}
			if disclosure.material.Method == "payload-key" {
				var encoded string
				key := "noteCiphertext"
				if disclosure.material.Kind == "credit" {
					key = "creditCiphertext"
					output.Role = "owner"
				}
				json.Unmarshal(fixture[key], &encoded)
				output.Ciphertext = common.FromHex(encoded)
			}
			result, e := disclosure.Verify(output)
			if e != nil {
				t.Fatal(e)
			}
			if result.Kind != disclosure.material.Kind {
				t.Fatal("disclosure kind mismatch")
			}
			serialized, e := json.Marshal(disclosure)
			if e != nil || strings.Contains(string(serialized), "payloadKey") || strings.Contains(string(serialized), "creditSharedKeyHash") {
				t.Fatal("implicit disclosure serialization leaked sensitive material")
			}
			if name == "notePayload" {
				fresh, e := viewer.DiscloseNote(output.Ciphertext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: output.ProtocolContext, OutputCommitment: core.BigIntToElement(output.OutputCommitment.Big())}, TokenAddress: output.TokenAddress})
				if e != nil {
					t.Fatal(e)
				}
				if fresh.Export().Data["payloadKey"] != disclosure.Export().Data["payloadKey"] {
					t.Fatal("payload key differs from TS SDK")
				}
			}
			output.OutputCommitment[0] ^= 1
			if _, e = disclosure.Verify(output); e == nil {
				t.Fatal("disclosure accepted a different output")
			}
		})
	}
}

func TestCreditOpeningSupportsFrozenUint128Usage(t *testing.T) {
	session, token, _ := testSession(t)
	wide := core.BigIntToElement(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(3)))
	state := core.CreditSpendingState{DailyLimit: core.ElementFromUint64(1), LastDay: core.ElementFromUint64(1), DailySpent: wide, LastMonth: core.ElementFromUint64(1), MonthlySpent: wide}
	leaf := seedLimitedCredit(t, session, token, state)
	snapshot, err := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	if err != nil {
		t.Fatal(err)
	}
	note := snapshot.Credits[creditKey(leaf, "owner")].State.Note
	disclosure, err := DisclosureFromCredit(note)
	if err != nil {
		t.Fatal(err)
	}
	output := DisclosureOutput{ProtocolContext: note.ProtocolContext, TokenAddress: note.TokenAddress, OutputCommitment: leaf, Role: "owner"}
	result, err := disclosure.Verify(output)
	if err != nil || result.Credit.DailySpent != wide || result.Credit.MonthlySpent != wide {
		t.Fatal("frozen uint128 opening rejected", err)
	}
	disclosure.material.Data["dailySpent"] = new(big.Int).Lsh(big.NewInt(1), 128).String()
	if _, err = disclosure.Verify(output); err == nil {
		t.Fatal("uint128 overflow opening accepted")
	}
}
