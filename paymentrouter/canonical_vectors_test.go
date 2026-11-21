package paymentrouter

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"os"
	"testing"
)

type canonicalVector struct {
	Name, IntentDigest, Signature, CoreCalldata string
	Call                                        struct{ To, Data, Value string }
	Account                                     *struct {
		Implementation, RouterCodeHash, Digest, Signature string
		Call                                              struct{ To, Data, Value string }
	}
}

func canonicalVectors(t *testing.T) (map[string]string, []canonicalVector) {
	t.Helper()
	b, e := os.ReadFile("testdata/payment-router-intent-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Addresses map[string]string
		Cases     []canonicalVector
	}
	if e = json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v.Addresses, v.Cases
}
func TestSharedCanonicalVectors(t *testing.T) {
	addresses, cases := canonicalVectors(t)
	for _, v := range cases {
		t.Run(v.Name, func(t *testing.T) {
			data := common.FromHex(v.Call.Data)
			p, e := DecodeRouterCall(data)
			if e != nil {
				t.Fatal(e)
			}
			b, e := EncodeRouterCall(p)
			if e != nil || !bytes.Equal(data, b) {
				t.Fatal("Router calldata mismatch", e)
			}
			h, e := HashIntent(bi(31337), common.HexToAddress(addresses["router"]), p.Intent)
			if e != nil || h.Hex() != v.IntentDigest {
				t.Fatal("intent digest mismatch", h, e)
			}
			if e = VerifySignature(h, common.FromHex(v.Signature), p.Intent.Payer); e != nil {
				t.Fatal(e)
			}
			ch, e := HashCoreRequest(p.Proof, p.Request)
			if e != nil || ch != p.Intent.CoreRequestHash {
				t.Fatal("Core hash mismatch", e)
			}
			rh, e := HashRoute(p.Route)
			if e != nil || rh != p.Intent.RouteHash {
				t.Fatal("route hash mismatch", e)
			}
			core, e := p.CoreCall(common.HexToAddress(addresses["core"]))
			if e != nil || hexutil.Encode(core.Data) != v.CoreCalldata {
				t.Fatal("Core calldata mismatch", e)
			}
			if v.Account != nil {
				a, e := DecodeAccountCall(common.FromHex(v.Account.Call.Data))
				if e != nil {
					t.Fatal(e)
				}
				b, e := EncodeAccountCall(a)
				if e != nil || hexutil.Encode(b) != v.Account.Call.Data {
					t.Fatal("account calldata mismatch", e)
				}
				h, e := HashNativePayment(bi(31337), p.Intent.Payer, common.HexToAddress(v.Account.Implementation), common.HexToAddress(addresses["router"]), common.HexToHash(v.Account.RouterCodeHash), a)
				if e != nil || h.Hex() != v.Account.Digest {
					t.Fatal("account digest mismatch", h, e)
				}
				if e = VerifySignature(h, a.Signature, p.Intent.Payer); e != nil {
					t.Fatal(e)
				}
				env := RouterEnvelope{To: p.Intent.Payer, Data: b, Value: "0"}
				first, e := EnvelopeHash(bi(31337), core.To, env)
				if e != nil {
					t.Fatal(e)
				}
				env.Data = append([]byte(nil), b...)
				env.Data[len(env.Data)-1] ^= 1
				second, e := EnvelopeHash(bi(31337), core.To, env)
				if e != nil || first == second {
					t.Fatal("envelope did not bind calldata", e)
				}
			}
		})
	}
}
func TestCanonicalValidationRejectsTampering(t *testing.T) {
	addresses, cases := canonicalVectors(t)
	d := deployment()
	d.CoreAddress = common.HexToAddress(addresses["core"])
	d.RouterAddress = common.HexToAddress(addresses["router"])
	d.WETHAddress = common.HexToAddress(addresses["weth"])
	for _, v := range cases {
		t.Run(v.Name, func(t *testing.T) {
			p, e := DecodeRouterCall(common.FromHex(v.Call.Data))
			if e != nil {
				t.Fatal(e)
			}
			if e = ValidateIntent(p, d, bi(1000), p.Intent.Nonce); e != nil {
				t.Fatal(e)
			}
			p.Intent.AmountInMaximum = bi(0)
			if e = ValidateIntent(p, d, bi(1000), p.Intent.Nonce); e == nil {
				t.Fatal("zero input accepted")
			}
			p, _ = DecodeRouterCall(common.FromHex(v.Call.Data))
			p.Intent.CoreRequestHash = common.Hash{}
			if e = ValidateIntent(p, d, bi(1000), p.Intent.Nonce); e == nil {
				t.Fatal("modified request hash accepted")
			}
			if _, e = DecodeRouterCall(append(common.FromHex(v.Call.Data), 0)); e == nil {
				t.Fatal("ABI suffix accepted")
			}
		})
	}
}
