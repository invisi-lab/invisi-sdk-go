package settlement

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSixOperationCanonicalVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/sdk-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []vector
	if err = json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 6 {
		t.Fatal("expected six canonical vectors")
	}
	want := map[string]bool{"transact": true, "shield": true, "createCredit": true, "creditDraw": true, "updateCredit": true, "revokeCredit": true}
	for _, v := range vectors {
		if !want[v.Method] {
			t.Fatalf("unexpected or duplicate operation %s", v.Method)
		}
		delete(want, v.Method)
		t.Run(v.Method, func(t *testing.T) {
			data := unhex(t, v.Data)
			request, err := DecodeSingle(data)
			if err != nil {
				t.Fatal(err)
			}
			if request.Method != v.Method {
				t.Fatalf("method %s", request.Method)
			}
			encoded, err := Encode(request)
			if err != nil || !bytes.Equal(data, encoded) {
				t.Fatalf("encode %v", err)
			}
			hash, err := RequestHash(data)
			if err != nil || hash.Hex() != v.RequestHash {
				t.Fatalf("SDK hash mismatch %s %v", hash, err)
			}
			variant, err := RequestHash(unhex(t, v.UnsignedVariant))
			if err != nil || variant != hash {
				t.Fatal("proof/funding signature altered intent", err)
			}
			digest, err := QuoteDigest(domain(v), v.Quote)
			if err != nil || digest.Hex() != v.QuoteDigest {
				t.Fatal("typed digest mismatch", err)
			}
			if err = VerifyQuote(domain(v), v.Quote, unhex(t, v.Signature), common.HexToAddress(v.Trust.Signer)); err != nil {
				t.Fatal(err)
			}
			if _, err = DecodeSingle(append(bytes.Clone(data), 0)); err == nil {
				t.Fatal("trailing calldata accepted")
			}
		})
	}
}

func operationVectors(t *testing.T) []vector {
	t.Helper()
	raw, e := os.ReadFile("testdata/sdk-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var out []vector
	if e = json.Unmarshal(raw, &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func TestOperationPublicFieldsAndLockProjection(t *testing.T) {
	expected := map[string]int{"transact": 2, "shield": 0, "createCredit": 0, "creditDraw": 1, "updateCredit": 1, "revokeCredit": 0}
	for _, v := range operationVectors(t) {
		t.Run(v.Method, func(t *testing.T) {
			data := unhex(t, v.Data)
			r, e := DecodeSingle(data)
			if e != nil {
				t.Fatal(e)
			}
			statement := r.Statement()
			if len(statement.Nullifiers) != expected[v.Method] || statement.FeeAmount.String() != "45" || statement.Deadline.String() != "2000000000" {
				t.Fatal("wrong projected statement", statement)
			}
			before := append([]byte(nil), r.Proof...)
			beforeSig := append([]byte(nil), r.Funding.Signature...)
			original, e := Hash(r)
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(before, r.Proof) || !bytes.Equal(beforeSig, r.Funding.Signature) {
				t.Fatal("Hash mutated caller request")
			}
			switch r.Method {
			case "shield":
				r.ShieldOutputs[0].Amount.Add(r.ShieldOutputs[0].Amount, common.Big1)
			case "createCredit":
				r.Creation.CreditCommitment[0] ^= 1
			case "creditDraw":
				r.Draw.GrantRefs[0].Slot++
			case "updateCredit":
				r.Update.Owner[0] ^= 1
			case "revokeCredit":
				r.Revocation.Owner[0] ^= 1
			case "transact":
				r.Params.WithdrawalAddress[0] ^= 1
			}
			changed, e := Hash(r)
			if e != nil || changed == original {
				t.Fatal("operation-specific business field not bound", e)
			}
			r, _ = DecodeSingle(data)
			r.Statement().FeeAmount.Add(r.Statement().FeeAmount, common.Big1)
			changed, e = Hash(r)
			if e != nil || changed == original {
				t.Fatal("fee not bound", e)
			}
		})
	}
}
func TestUnsupportedOperations(t *testing.T) {
	raw, e := os.ReadFile("testdata/unsupported.json")
	if e != nil {
		t.Fatal(e)
	}
	var unsupported map[string]string
	if e = json.Unmarshal(raw, &unsupported); e != nil {
		t.Fatal(e)
	}
	for name, data := range unsupported {
		t.Run(name, func(t *testing.T) {
			if _, e = DecodeSingle(common.FromHex(data)); e == nil {
				t.Fatal("unsupported calldata accepted")
			}
			if _, e = RequestHash(common.FromHex(data)); e == nil {
				t.Fatal("unsupported intent hash accepted")
			}
		})
	}
}

func TestPublicFundedZeroInputSDKVector(t *testing.T) {
	raw, e := os.ReadFile("testdata/sdk-public-funding-vector.json")
	if e != nil {
		t.Fatal(e)
	}
	var v vector
	if e = json.Unmarshal(raw, &v); e != nil {
		t.Fatal(e)
	}
	request, e := DecodeSingle(common.FromHex(v.Data))
	if e != nil {
		t.Fatal(e)
	}
	if request.Method != "transact" || len(request.Statement().Nullifiers) != 0 || request.Funding.Amount.Sign() <= 0 {
		t.Fatal("unexpected public-funded fixture")
	}
	hash, e := RequestHash(common.FromHex(v.Data))
	if e != nil || hash.Hex() != v.RequestHash {
		t.Fatal("public funded intent differs from actual SDK", e)
	}
	variant, e := RequestHash(common.FromHex(v.UnsignedVariant))
	if e != nil || variant != hash {
		t.Fatal("public funded proof/signature altered intent", e)
	}
	if e = VerifyQuote(domain(v), v.Quote, common.FromHex(v.Signature), common.HexToAddress(v.Trust.Signer)); e != nil {
		t.Fatal(e)
	}
}
