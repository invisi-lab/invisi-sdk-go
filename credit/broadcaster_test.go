package credit

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

type creditRoundTrip func(*http.Request) (*http.Response, error)

func (f creditRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func creditBroadcaster(t *testing.T, c Context, signer *ecdsa.PrivateKey, collector common.Address, policies ...common.Hash) *broadcaster.Client {
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
		case "broadcaster_estimateQuote":
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
			if len(policies) > 0 {
				bound, e := provenance.ProvenanceIntentHash(hash.Hex())
				if e != nil {
					t.Fatal(e)
				}
				hash = common.HexToHash(bound)
			}
			q := broadcaster.Quote{Token: intent.Token, FeeRecipient: intent.FeeRecipient, FeeAmount: intent.FeeAmount.String(), Deadline: intent.Deadline.String(), MaxGas: "2000000", RequestHash: hash, Execution: "direct"}
			typed, err := broadcaster.QuoteTypedData(q, broadcaster.Trust{ChainID: primitives.ElementToBigInt(c.ChainID), Protocol: common.BigToAddress(primitives.ElementToBigInt(c.ProtocolAddress)), Signer: crypto.PubkeyToAddress(signer.PublicKey)})
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
			signed := broadcaster.SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(signature)}
			if len(policies) > 0 {
				signed.ProvenanceRequired = true
				signed.RequestHashFormat = broadcaster.ProvenanceRequestHashFormat
			}
			result = signed
		case "broadcaster_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "chainId": primitives.ElementToBigInt(c.ChainID).String(), "protocol": common.BigToAddress(primitives.ElementToBigInt(c.ProtocolAddress)).Hex(), "sender": crypto.PubkeyToAddress(signer.PublicKey).Hex(), "feeRecipient": collector.Hex(), "submissionHalted": false, "provenance": map[string]any{"enabled": len(policies) > 0, "requestHashFormat": broadcaster.ProvenanceRequestHashFormat, "trackingRequired": true, "trackingSupported": true}}
		case "broadcaster_getTransactionStatus":
			result = broadcaster.Status{ID: request.Params[0]["id"].(string), State: "uncertain"}
		case "broadcaster_submitTransaction":
			result = broadcaster.Status{ID: "server-request-id", State: "submitted", TransactionHash: common.HexToHash("0x1234").Hex()}
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
	if quoted.Create.FeeRecipient != collector || quoted.Create.FeeAmount.Cmp(big.NewInt(1)) != 0 || quoted.Create.CreditCommitment != p.CreditCommitment || quoted.CreatedState.CreditSharedKey != s.CreditSharedKey {
		t.Fatal("quote changed grant or omitted fee")
	}
}

func TestCreditRequiredPolicyQuoteWithoutSourceProver(t *testing.T) {
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
	client := creditBroadcaster(t, c, key, collector, common.HexToHash("0x77"))
	if err = prepared.WithBroadcaster(context.Background(), client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	quoted, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if quoted.Create.FeeRecipient != collector || quoted.Create.FeeAmount.Cmp(big.NewInt(1)) != 0 || quoted.Create.CreditCommitment != p.CreditCommitment || quoted.CreatedState.CreditSharedKey != s.CreditSharedKey {
		t.Fatal("quote changed grant or omitted fee")
	}
}

func TestCreditKnownSubmissionQueriesWithoutFreshProof(t *testing.T) {
	s, _ := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	clue := make([]byte, 68)
	clue[64] = 24
	commitment, _ := s.Commitment()
	prepared, err := PrepareCreate(c, CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}, []protocol.Delivery{{Ciphertext: []byte{0x82}, Clue: clue}, {Ciphertext: []byte{0x82}, Clue: clue}}, s)
	if err != nil {
		t.Fatal(err)
	}
	key, _ := crypto.GenerateKey()
	client := creditBroadcaster(t, c, key, common.HexToAddress("0x1234"), common.HexToHash("0x77"))
	if err = prepared.WithBroadcaster(context.Background(), client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	prepared.submission = Submission{RequestID: "original", State: "uncertain"}
	status, err := prepared.SubmitBroadcaster(context.Background(), client)
	if err != nil || status.State != "uncertain" {
		t.Fatal("known externalized request needed a new proof/send", status, err)
	}
}
