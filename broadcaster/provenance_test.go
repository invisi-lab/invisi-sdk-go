package broadcaster

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"testing"
)

func TestProvenanceQuotePolicyBinding(t *testing.T) {
	data, trust, key := intent(t)
	legacy := signed(t, data, trust, key)
	policy := common.HexToHash("0x77")
	bound, e := provenance.ProvenanceIntentHash(legacy.Quote.RequestHash.Hex(), policy.Hex())
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
	s := ProvenanceSignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(signature), ProvenancePolicyID: &policy, RequestHashFormat: ProvenanceRequestHashFormat}
	if e := VerifyProvenanceQuote(s, trust, data, 0, policy); e != nil {
		t.Fatal(e)
	}
	if VerifyProvenanceQuote(s, trust, data, 0, common.HexToHash("0x78")) == nil {
		t.Fatal("different policy accepted")
	}
	s.ProvenancePolicyID = nil
	if VerifyProvenanceQuote(s, trust, data, 0, policy) == nil {
		t.Fatal("missing policy accepted")
	}
	s.ProvenancePolicyID = &policy
	if VerifyQuote(SignedQuote{Quote: s.Quote, Signature: s.Signature}, trust, data, 0) == nil {
		t.Fatal("bound quote downgraded to legacy")
	}
	wire, _ := json.Marshal(s)
	var decoded SignedQuote
	if json.Unmarshal(wire, &decoded) == nil {
		t.Fatal("legacy quote silently dropped policy metadata")
	}
}
