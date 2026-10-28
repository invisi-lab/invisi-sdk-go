package core

import (
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"os"
	"testing"
)

type protocolVector struct {
	Name     string            `json:"name"`
	Inputs   []string          `json:"inputs"`
	Expected json.RawMessage   `json:"expected"`
	Context  map[string]string `json:"context"`
	Previous map[string]string `json:"previous"`
	Recovery map[string]string `json:"recovery"`
}

func vectorElement(t *testing.T, s string) fr.Element {
	t.Helper()
	var e fr.Element
	if _, err := e.SetString(s); err != nil {
		t.Fatalf("invalid fixture element %q: %v", s, err)
	}
	return e
}
func loadProtocolVectors(t *testing.T) []protocolVector {
	t.Helper()
	data, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Vectors []protocolVector `json:"vectors"`
	}
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Vectors) != 33 {
		t.Fatalf("expected 33 protocol vectors, got %d", len(file.Vectors))
	}
	return file.Vectors
}
func TestCurrentProtocolVectors(t *testing.T) {
	for _, v := range loadProtocolVectors(t) {
		if v.Name == "RecoverCreditRemainder" {
			continue
		}
		t.Run(v.Name, func(t *testing.T) {
			p := make([]fr.Element, len(v.Inputs))
			for i, s := range v.Inputs {
				p[i] = vectorElement(t, s)
			}
			var got fr.Element
			var err error
			switch v.Name {
			case "DeriveNoteSecret":
				got, err = DeriveNoteSecret(p[0], p[1])
			case "ComputeIdentityCommitment":
				got, err = ComputeIdentityCommitment(p[0], p[1])
			case "ComputeNoteCommitment":
				got, err = ComputeNoteCommitment(p[0], p[1], p[2], p[3], p[4])
			case "ComputeChangeCommitment":
				got, err = ComputeChangeCommitment(p[0], p[1], p[2], p[3], p[4], p[5])
			case "ComputeNoteNullifier":
				got, err = ComputeNoteNullifier(p[0], p[1], p[2], p[3], p[4], p[5])
			case "ComputeCreditSharedKeyHash":
				got, err = ComputeCreditSharedKeyHash(p[0])
			case "ComputeCreditCommitment":
				got, err = ComputeCreditCommitment(p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7])
			case "ComputeCreditNullifier":
				got, err = ComputeCreditNullifier(p[0], p[1], p[2], p[3], p[4], p[5])
			case "DeriveRemainderNonce":
				got, err = DeriveRemainderNonce(p[0], p[1], p[2], p[3], p[4])
			case "ComputeCreditRecoveryPad":
				got, err = ComputeCreditRecoveryPad(p[0], p[1], p[2], p[3], p[4])
			default:
				t.Fatalf("unknown vector %s", v.Name)
			}
			if err != nil {
				t.Fatal(err)
			}
			var expected string
			if err = json.Unmarshal(v.Expected, &expected); err != nil {
				t.Fatal(err)
			}
			want := vectorElement(t, expected)
			if got != want {
				t.Fatalf("got %s want %s", ElementToHex(got), ElementToHex(want))
			}
		})
	}
}
func creditFixture(t *testing.T, m map[string]string) CreditNote {
	return CreditNote{SpenderIdentityCommitment: vectorElement(t, m["spenderIdentityCommitment"]), OwnerIdentityCommitment: vectorElement(t, m["ownerIdentityCommitment"]), TokenAddress: vectorElement(t, m["tokenAddress"]), Amount: vectorElement(t, m["amount"]), Nonce: vectorElement(t, m["nonce"]), Expiry: vectorElement(t, m["expiry"]), CreditSharedKey: vectorElement(t, m["creditSharedKey"])}
}
func TestRecoverCreditRemainderProtocolVectors(t *testing.T) {
	for _, v := range loadProtocolVectors(t) {
		if v.Name != "RecoverCreditRemainder" {
			continue
		}
		ctx := ProtocolContext{ChainID: vectorElement(t, v.Context["chainId"]), ProtocolAddress: vectorElement(t, v.Context["protocolAddress"])}
		previous := creditFixture(t, v.Previous)
		recovery := CreditRemainderRecovery{CreditNullifier: vectorElement(t, v.Recovery["creditNullifier"]), RemainderCommitment: vectorElement(t, v.Recovery["remainderCommitment"]), RecoveryAmount: vectorElement(t, v.Recovery["recoveryAmount"])}
		var expected map[string]string
		if err := json.Unmarshal(v.Expected, &expected); err != nil {
			t.Fatal(err)
		}
		got, err := RecoverCreditRemainder(ctx, previous, recovery)
		if err != nil {
			t.Fatal(err)
		}
		want := creditFixture(t, expected)
		if got == nil || *got != want {
			t.Fatalf("recovered note mismatch: %+v", got)
		}
		full := recovery
		full.RemainderCommitment = fr.Element{}
		full.RecoveryAmount = fr.Element{}
		if got, err := RecoverCreditRemainder(ctx, previous, full); err != nil || got != nil {
			t.Fatalf("full draw: %v %v", got, err)
		}
		cases := map[string]CreditRemainderRecovery{"nullifier": recovery, "commitment": recovery, "amount": recovery, "full draw amount": full}
		bad := cases["nullifier"]
		bad.CreditNullifier = ElementFromUint64(1)
		cases["nullifier"] = bad
		bad = cases["commitment"]
		bad.RemainderCommitment = ElementFromUint64(1)
		cases["commitment"] = bad
		bad = cases["amount"]
		bad.RecoveryAmount = ElementFromUint64(1)
		cases["amount"] = bad
		bad = cases["full draw amount"]
		bad.RecoveryAmount = ElementFromUint64(1)
		cases["full draw amount"] = bad
		for name, bad := range cases {
			t.Run(name, func(t *testing.T) {
				if _, err := RecoverCreditRemainder(ctx, previous, bad); err == nil {
					t.Fatal("accepted invalid recovery")
				}
			})
		}
		wrongContext := ctx
		wrongContext.ProtocolAddress = ElementFromUint64(1)
		if _, err := RecoverCreditRemainder(wrongContext, previous, recovery); err == nil {
			t.Fatal("accepted different deployment")
		}
	}
}
