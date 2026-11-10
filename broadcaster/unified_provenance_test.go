package broadcaster

import (
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestRequiredPolicyUsesNormalClient(t *testing.T) {
	source, e := NewProvenanceClient(Configuration{Endpoint: "https://relay.example", QuoteSigner: common.HexToAddress("0x11"), FeeRecipient: common.HexToAddress("0x22")}, common.HexToHash("0x77"))
	if e != nil {
		t.Fatal(e)
	}
	var normal *Client = source
	if normal.RequiredProvenancePolicy() != common.HexToHash("0x77") {
		t.Fatal("required source policy was not carried by normal client")
	}
}
