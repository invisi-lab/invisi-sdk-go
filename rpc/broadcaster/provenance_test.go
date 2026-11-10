package broadcaster

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
)

func TestProvenanceQuoteRequirementBinding(t *testing.T) {
	data, trust, key := intent(t)
	legacy := signed(t, data, trust, key)
	bound, e := provenance.ProvenanceIntentHash(legacy.Quote.RequestHash.Hex())
	if e != nil {
		t.Fatal(e)
	}
	q := legacy.Quote
	q.RequestHash = common.HexToHash(bound)
	td, e := QuoteTypedData(q, trust)
	if e != nil {
		t.Fatal(e)
	}
	digest, _, e := apitypes.TypedDataAndHash(td)
	if e != nil {
		t.Fatal(e)
	}
	signature, e := crypto.Sign(digest, key)
	if e != nil {
		t.Fatal(e)
	}
	s := SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(signature), ProvenanceRequired: true, RequestHashFormat: ProvenanceRequestHashFormat}
	if e := VerifyQuote(s, trust, data, 0); e != nil {
		t.Fatal(e)
	}
	s.ProvenanceRequired = false
	if VerifyQuote(s, trust, data, 0) == nil {
		t.Fatal("missing requirement accepted")
	}
	s.ProvenanceRequired = true
	if VerifyQuote(SignedQuote{Quote: s.Quote, Signature: s.Signature}, trust, data, 0) == nil {
		t.Fatal("bound quote downgraded to legacy")
	}
	wire, _ := json.Marshal(s)
	var decoded SignedQuote
	if json.Unmarshal(wire, &decoded) != nil || !decoded.ProvenanceRequired {
		t.Fatal("quote policy metadata lost")
	}
	if VerifyQuote(decoded, trust, data, 0) != nil {
		t.Fatal("required provenance quote failed verification")
	}
}

func TestQuoteRejectsUnknownFieldsAndFalseRequirement(t *testing.T) {
	for _, raw := range []string{`{"quote":{},"signature":"0x","unexpectedField":"1"}`, `{"quote":{},"signature":"0x","provenanceRequired":false}`} {
		var quote SignedQuote
		if json.Unmarshal([]byte(raw), &quote) == nil {
			t.Fatalf("unknown or malformed quote metadata accepted: %s", raw)
		}
	}
}
