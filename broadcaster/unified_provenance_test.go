package broadcaster

import (
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestClientDoesNotRequireLocalProvenanceConfiguration(t *testing.T) {
	if _, e := NewClient(Configuration{Endpoint: "https://relay.example", QuoteSigner: common.HexToAddress("0x11"), FeeRecipient: common.HexToAddress("0x22")}); e != nil {
		t.Fatal(e)
	}
}
