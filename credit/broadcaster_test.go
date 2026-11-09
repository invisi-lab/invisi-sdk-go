package credit

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
)

type creditRoundTrip func(*http.Request) (*http.Response, error)

func (f creditRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func creditBroadcaster(t *testing.T, c Context, signer *ecdsa.PrivateKey, collector common.Address) *broadcaster.Client {
	t.Helper()
	transport := creditRoundTrip(func(r *http.Request) (*http.Response, error) {
		var request struct {
			ID     json.RawMessage
			Method string
			Params []map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		var result any
		switch request.Method {
		case "broadcaster_getFee":
			result = broadcaster.Fee{FeeRecipient: collector, FeeAmount: "1", MaxGas: "2000000"}
		case "broadcaster_getQuote":
			data := common.FromHex(request.Params[0]["data"].(string))
			intent, err := broadcaster.DecodeIntent(data)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := broadcaster.RequestHash(data)
			if err != nil {
				t.Fatal(err)
			}
			q := broadcaster.Quote{Token: intent.Token, FeeRecipient: intent.FeeRecipient, FeeAmount: intent.FeeAmount.String(), Deadline: intent.Deadline.String(), MaxGas: "2000000", RequestHash: hash}
			typed, err := broadcaster.QuoteTypedData(q, broadcaster.Trust{ChainID: core.ElementToBigInt(c.ChainID), Protocol: common.BigToAddress(core.ElementToBigInt(c.ProtocolAddress)), Signer: crypto.PubkeyToAddress(signer.PublicKey)})
			if err != nil {
				t.Fatal(err)
			}
			digest, _, err := apitypes.TypedDataAndHash(typed)
			if err != nil {
				t.Fatal(err)
			}
			signature, err := crypto.Sign(digest, signer)
			if err != nil {
				t.Fatal(err)
			}
			result = broadcaster.SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(signature)}
		case "broadcaster_submitTransaction":
			result = broadcaster.Status{ID: request.Params[0]["idempotencyKey"].(string), State: "submitted", TransactionHash: common.HexToHash("0x1234").Hex()}
		default:
			t.Fatalf("unexpected method %s", request.Method)
		}
		raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
	})
	client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: "https://operator.example", QuoteSigner: crypto.PubkeyToAddress(signer.PublicKey), FeeRecipient: collector}, broadcaster.Options{HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func TestCreditQuoteFreezesFeeBeforeAuthorization(t *testing.T) {
	s, f := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	commitment, _ := s.Commitment()
	p := CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	prepared, err := PrepareCreate(c, p, ds, s)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := prepared.ExportBroadcasterIntent()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := broadcaster.DecodeIntent(intent)
	if err != nil || decoded.Method != "createCredit" || decoded.FeeAmount.Sign() != 0 {
		t.Fatal("unsigned Credit intent", err)
	}
	key, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	collector := common.HexToAddress("0x1234")
	client := creditBroadcaster(t, c, key, collector)
	if err = prepared.WithBroadcaster(context.Background(), client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	quoted, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if quoted.Create.FeeRecipient != collector || quoted.Create.FeeAmount.Cmp(big.NewInt(1)) != 0 || quoted.Create.CreditCommitment != p.CreditCommitment || quoted.CreatedState.SharedKey != s.SharedKey {
		t.Fatal("quote changed grant or omitted fee")
	}
}
