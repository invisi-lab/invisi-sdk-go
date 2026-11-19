package invisi

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/ipp"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
)

type quoteTransport func(*http.Request) (*http.Response, error)

func (f quoteTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestIPPNoFeeThenBroadcasterQuotePreservesFrozenNet(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	request := fixture(t, "credit")
	request.CreatedAt = now.Format(time.RFC3339)
	request.ExpiresAt = now.Add(5 * time.Minute).Format(time.RFC3339)
	request.Credit.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
	request.Credit.InitialPayment = &ipp.InitialPaymentTerms{Amount: "5"}
	request.Credit.AvailableCredit = "10"
	owner, e := primitives.DeriveWallet(common.LeftPadBytes([]byte{1}, 32))
	if e != nil {
		t.Fatal(e)
	}
	address, e := primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: owner.OwnerAddress, ViewingPublicKey: owner.ViewingPublicKey, ClueKey: owner.ClueKey})
	if e != nil {
		t.Fatal(e)
	}
	deployment := protocolrpc.DeploymentContext{ChainID: 8453, ProtocolAddress: common.HexToAddress(request.Credit.ProtocolAddress)}
	builder := credit.PreparationBuilder{
		Frame: func(_ context.Context, token common.Address, deadline uint64) (credit.Context, uint64, error) {
			return credit.Context{ProtocolContext: primitives.ProtocolContext{ChainID: primitives.ElementFromUint64(8453), ProtocolAddress: primitives.BigIntToElement(deployment.ProtocolAddress.Big())}, TokenAddress: token, Timestamp: uint64(now.Unix())}, deadline, nil
		},
		Owner: func() (string, error) { return address, nil },
		Resolve: func(_ context.Context, a string) (primitives.ReceivingAddress, error) {
			return primitives.ResolveAddress(a, nil)
		},
		Allocation: func(context.Context, credit.Context) (uint64, uint16, error) { return 1, 0, nil },
	}
	key, _ := ethcrypto.ToECDSA(common.LeftPadBytes([]byte{7}, 32))
	recipient := common.HexToAddress("44")
	for _, test := range []struct {
		name     string
		fee, cap int64
		quota    string
		fail     bool
	}{{"fee one", 1, 1, "10", false}, {"fee cap", 2, 1, "10", true}, {"quota", 1, 2, "5", true}} {
		t.Run(test.name, func(t *testing.T) {
			request.Credit.SingleLimit = test.quota
			prepared, e := PrepareRequest(ctx, request, Target[*credit.Prepared]{Context: deployment, PrepareCreateCredit: builder.Create}, Options{Now: func() time.Time { return now }})
			if e != nil {
				t.Fatal(e)
			}
			original, _ := prepared.ExportIntent()
			raw, _ := json.Marshal(original)
			var restored credit.Intent
			if e = json.Unmarshal(raw, &restored); e != nil {
				t.Fatal(e)
			}
			prepared, e = credit.Restore(ctx, restored, nil, nil)
			if e != nil {
				t.Fatal(e)
			}
			transport := quoteTransport(func(r *http.Request) (*http.Response, error) {
				var q struct {
					ID     json.RawMessage
					Method string
					Params []map[string]any
				}
				if e := json.NewDecoder(r.Body).Decode(&q); e != nil {
					t.Fatal(e)
				}
				var result any
				switch q.Method {
				case "broadcaster_estimateQuote":
					result = broadcaster.Fee{FeeRecipient: recipient, FeeAmount: big.NewInt(test.fee).String(), MaxGas: "2000000"}
				case "broadcaster_getQuote":
					data := common.FromHex(q.Params[0]["data"].(string))
					intent, e := broadcaster.DecodeIntent(data)
					if e != nil {
						t.Fatal(e)
					}
					hash, e := broadcaster.RequestHash(data)
					if e != nil {
						t.Fatal(e)
					}
					quote := broadcaster.Quote{Token: intent.Token, FeeRecipient: recipient, FeeAmount: big.NewInt(test.fee).String(), Deadline: intent.Deadline.String(), MaxGas: "2000000", RequestHash: hash, Execution: "direct"}
					typed, e := broadcaster.QuoteTypedData(quote, broadcaster.Trust{ChainID: big.NewInt(8453), Protocol: deployment.ProtocolAddress, Signer: ethcrypto.PubkeyToAddress(key.PublicKey)})
					if e != nil {
						t.Fatal(e)
					}
					digest, _, e := apitypes.TypedDataAndHash(typed)
					if e != nil {
						t.Fatal(e)
					}
					sig, e := ethcrypto.Sign(digest, key)
					if e != nil {
						t.Fatal(e)
					}
					result = broadcaster.SignedQuote{Quote: quote, Signature: "0x" + common.Bytes2Hex(sig)}
				default:
					t.Fatalf("unexpected RPC %s", q.Method)
				}
				raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": result})
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
			})
			client, e := broadcaster.NewClient(broadcaster.Configuration{Endpoint: "https://relay.example", QuoteSigner: ethcrypto.PubkeyToAddress(key.PublicKey), FeeRecipient: recipient}, broadcaster.Options{HTTPClient: &http.Client{Transport: transport}})
			if e != nil {
				t.Fatal(e)
			}
			e = client.Quote(ctx, prepared, credit.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(test.cap)})
			if test.fail {
				if e == nil {
					t.Fatal("accepted fee or quota overflow")
				}
				unchanged, _ := prepared.ExportIntent()
				if unchanged.Create.TotalAmount.Int64() != 5 || unchanged.Broadcaster != nil {
					t.Fatal("failed quote modified original")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			got, _ := prepared.ExportIntent()
			if got.Create.TotalAmount.Int64() != 6 || got.Create.FeeAmount.Int64() != 1 || primitives.ElementToBigInt(got.OutputNote.Amount).Int64() != 5 || primitives.ElementToBigInt(got.CreatedState.AvailableCredit).Int64() != 4 {
				t.Fatal("frozen net lost during quote")
			}
			if request.Credit.InitialPayment.Amount != "5" {
				t.Fatal("merchant request mutated")
			}
			if e = prepared.CheckTimestamp(uint64((now.Unix()/86400 + 1) * 86400)); e == nil {
				t.Fatal("accepted next UTC period")
			}
		})
	}
}
