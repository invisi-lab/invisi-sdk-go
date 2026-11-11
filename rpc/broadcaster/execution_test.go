package broadcaster

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestExecutionSeventhSignedField(t *testing.T) {
	q := Quote{Execution: "direct"}
	td, err := QuoteTypedData(q, Trust{ChainID: big.NewInt(1), Protocol: common.HexToAddress("0x1"), Signer: common.HexToAddress("0x2")})
	if err != nil {
		t.Fatal(err)
	}
	fields := td.Types["Quote"]
	if len(fields) != 7 || fields[6].Name != "execution" || td.Message["execution"] != "direct" {
		t.Fatal("execution must be seventh signed field")
	}
	q.Execution = ""
	if _, err = QuoteTypedData(q, Trust{ChainID: big.NewInt(1), Protocol: common.HexToAddress("0x1"), Signer: common.HexToAddress("0x2")}); err == nil {
		t.Fatal("missing execution accepted")
	}
}
func TestAdditionalStatuses(t *testing.T) {
	for _, s := range []string{"queued", "proving", "rejected", "expired"} {
		if err := checkStatus(Status{ID: "a", State: s}, ""); err != nil {
			t.Fatal(err)
		}
	}
}

func TestQuoteRejectsUnknownAndMissingExecution(t *testing.T) {
	data, trust, key := intent(t)
	signed := signed(t, data, trust, key)
	raw, err := json.Marshal(signed.Quote)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["unsignedExecution"] = "aggregate"
	raw, _ = json.Marshal(fields)
	var q Quote
	if json.Unmarshal(raw, &q) == nil {
		t.Fatal("unknown unsigned quote field accepted")
	}
	delete(fields, "unsignedExecution")
	delete(fields, "execution")
	raw, _ = json.Marshal(fields)
	if json.Unmarshal(raw, &q) == nil {
		t.Fatal("missing execution accepted")
	}
	signed.Quote.Execution = "aggregate"
	if VerifyQuote(signed, trust, data, 0) == nil {
		t.Fatal("execution tampering accepted")
	}
}

func TestFrozenEnvelopeRejectsUnknownFields(t *testing.T) {
	for _, raw := range []string{`null`, `{"privateWitness":[]}`, `{} {}`} {
		var proof MemberProof
		if json.Unmarshal([]byte(raw), &proof) == nil {
			t.Fatal("invalid public member accepted", raw)
		}
	}
	var submission Submission
	raw := []byte(`{"data":"0x","quote":{"token":"0x0000000000000000000000000000000000000000","feeRecipient":"0x0000000000000000000000000000000000000000","feeAmount":"1","deadline":"1","maxGas":"1","requestHash":"0x0000000000000000000000000000000000000000000000000000000000000000","execution":"aggregate"},"signature":"0x","memberProof":{"privateWitness":[]}}`)
	if json.Unmarshal(raw, &submission) == nil {
		t.Fatal("restored unknown member field accepted")
	}
}
