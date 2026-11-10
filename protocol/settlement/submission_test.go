package settlement

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSubmissionIDCanonicalScopeAndSemanticVariants(t *testing.T) {
	v := fixture(t)
	d := domain(v)
	id, err := SubmissionID(d.ChainID, d.Protocol, unhex(t, v.Data))
	if err != nil {
		t.Fatal(err)
	}
	if id.Hex() != "0x5d0bed03bbca1d5de0bd643f2532cc8a57c547c5f014051af3f9d495fe9bcbcf" {
		t.Fatalf("cross-language submission vector changed: %s", id)
	}
	same, err := SubmissionID(d.ChainID, d.Protocol, unhex(t, v.UnsignedVariant))
	if err != nil || same != id {
		t.Fatalf("authorization changed identity: %s %v", same, err)
	}
	for _, mutate := range []func(*Request){
		func(r *Request) { r.Params.FeeAmount.Add(r.Params.FeeAmount, big.NewInt(1)) },
		func(r *Request) { r.Params.Deadline.Add(r.Params.Deadline, big.NewInt(1)) },
		func(r *Request) { r.Params.Nullifiers[0][0] ^= 1 },
		func(r *Request) { r.Params.NoteCommitments[0][0] ^= 1 },
	} {
		r, err := DecodeSingle(unhex(t, v.Data))
		if err != nil {
			t.Fatal(err)
		}
		mutate(r)
		encoded, err := Encode(r)
		if err != nil {
			t.Fatal(err)
		}
		changed, err := SubmissionID(d.ChainID, d.Protocol, encoded)
		if err != nil || changed == id {
			t.Fatalf("semantic change lost: %s %v", changed, err)
		}
	}
	for _, scope := range []QuoteDomain{{ChainID: new(big.Int).Add(d.ChainID, big.NewInt(1)), Protocol: d.Protocol}, {ChainID: d.ChainID, Protocol: common.HexToAddress("0x9999999999999999999999999999999999999999")}} {
		changed, err := SubmissionID(scope.ChainID, scope.Protocol, unhex(t, v.Data))
		if err != nil || changed == id {
			t.Fatal("scope not bound", err)
		}
	}
}

func TestSubmissionIDAuthorizationVariantsAcrossOperations(t *testing.T) {
	raw, err := os.ReadFile("testdata/sdk-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []vector
	if err = json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, v := range vectors {
		d := domain(v)
		original, err := SubmissionID(d.ChainID, d.Protocol, unhex(t, v.Data))
		if err != nil {
			t.Fatal(err)
		}
		variant, err := SubmissionID(d.ChainID, d.Protocol, unhex(t, v.UnsignedVariant))
		if err != nil || variant != original {
			t.Fatalf("%s proof/signature altered ID: %v", v.Method, err)
		}
	}
}
