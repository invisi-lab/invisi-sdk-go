package settlement

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type vector struct {
	Method                                                                 string
	Data, RequestHash, UnsignedVariant, Signature, QuoteDigest, PrivateKey string
	Quote                                                                  Quote
	Trust                                                                  struct {
		ChainID          string `json:"chainId"`
		Protocol, Signer string
	}
}

func fixture(t *testing.T) vector {
	t.Helper()
	b, e := os.ReadFile("testdata/sdk-vector.json")
	if e != nil {
		t.Fatal(e)
	}
	var v vector
	if e = json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, e := hex.DecodeString(s[2:])
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func domain(v vector) QuoteDomain {
	chain, _ := new(big.Int).SetString(v.Trust.ChainID, 10)
	return QuoteDomain{ChainID: chain, Protocol: common.HexToAddress(v.Trust.Protocol)}
}
func TestActualSDKIntentVector(t *testing.T) {
	v := fixture(t)
	data := unhex(t, v.Data)
	r, e := DecodeSingle(data)
	if e != nil {
		t.Fatal(e)
	}
	if r.Params.FeeAmount.Cmp(big.NewInt(45)) != 0 || len(r.Params.Sources) != 2 || len(r.Deliveries) != 1 || r.Funding.SigNonce.Cmp(big.NewInt(8)) != 0 {
		t.Fatal("decoded request differs")
	}
	encoded, e := Encode(r)
	if e != nil || !bytes.Equal(encoded, data) {
		t.Fatalf("round trip: %v", e)
	}
	h, e := RequestHash(data)
	if e != nil || h.Hex() != v.RequestHash {
		t.Fatalf("hash %s, want %s (%v)", h, v.RequestHash, e)
	}
	h2, e := RequestHash(unhex(t, v.UnsignedVariant))
	if e != nil || h2 != h {
		t.Fatalf("proof/signature incorrectly bound: %v", e)
	}
}
func TestRejectNoncanonicalCalldata(t *testing.T) {
	v := fixture(t)
	b := unhex(t, v.Data)
	for _, bad := range [][]byte{b[:3], append(append([]byte{}, b...), 0), append(append([]byte{}, b...), make([]byte, 32)...), append([]byte{0, 0, 0, 0}, b[4:]...)} {
		if _, e := DecodeSingle(bad); e == nil {
			t.Fatal("accepted noncanonical data")
		}
	}
}
func TestActualEthersQuoteVector(t *testing.T) {
	v := fixture(t)
	d := domain(v)
	digest, e := QuoteDigest(d, v.Quote)
	if e != nil || digest.Hex() != v.QuoteDigest {
		t.Fatalf("digest %s (%v)", digest, e)
	}
	key, e := crypto.HexToECDSA(v.PrivateKey[2:])
	if e != nil {
		t.Fatal(e)
	}
	sig, e := SignQuote(d, v.Quote, key)
	if e != nil || common.Bytes2Hex(sig) != v.Signature[2:] {
		t.Fatalf("signature mismatch: %v", e)
	}
	if e = VerifyQuote(d, v.Quote, sig, common.HexToAddress(v.Trust.Signer)); e != nil {
		t.Fatal(e)
	}
	bad := v.Quote
	bad.FeeAmount = "46"
	if VerifyQuote(d, bad, sig, common.HexToAddress(v.Trust.Signer)) == nil {
		t.Fatal("accepted changed fee")
	}
	d.ChainID = big.NewInt(1)
	if VerifyQuote(d, v.Quote, sig, common.HexToAddress(v.Trust.Signer)) == nil {
		t.Fatal("accepted foreign chain")
	}
}

func TestIntentHashBindsAllPublicFields(t *testing.T) {
	v := fixture(t)
	data := unhex(t, v.Data)
	original, _ := RequestHash(data)
	mutations := []func(*Request){
		func(r *Request) { r.Params.Token[0] ^= 1 }, func(r *Request) { r.Params.Sources[0].TreeId.Add(r.Params.Sources[0].TreeId, big.NewInt(1)) }, func(r *Request) { r.Params.Sources[0].MerkleRoot[0] ^= 1 }, func(r *Request) { r.Params.Nullifiers[0][0] ^= 1 }, func(r *Request) { r.Params.NoteCommitments[0][0] ^= 1 },
		func(r *Request) { r.Params.WithdrawalAddress[0] ^= 1 }, func(r *Request) { r.Params.WithdrawalAmount.Add(r.Params.WithdrawalAmount, big.NewInt(1)) }, func(r *Request) { r.Params.Deadline.Add(r.Params.Deadline, big.NewInt(1)) }, func(r *Request) { r.Params.FeeRecipient[0] ^= 1 }, func(r *Request) { r.Params.FeeAmount.Add(r.Params.FeeAmount, big.NewInt(1)) },
		func(r *Request) { r.Deliveries[0].Ciphertext[0] ^= 1 }, func(r *Request) { r.Deliveries[0].Clue[0] ^= 1 }, func(r *Request) { r.Funding.Payer[0] ^= 1 }, func(r *Request) { r.Funding.Amount.Add(r.Funding.Amount, big.NewInt(1)) }, func(r *Request) { r.Funding.SigNonce.Add(r.Funding.SigNonce, big.NewInt(1)) }, func(r *Request) { r.Funding.Deadline.Add(r.Funding.Deadline, big.NewInt(1)) },
		func(r *Request) { r.Funding.Permit.Deadline.Add(r.Funding.Permit.Deadline, big.NewInt(1)) }, func(r *Request) { r.Funding.Permit.V ^= 1 }, func(r *Request) { r.Funding.Permit.R[0] ^= 1 }, func(r *Request) { r.Funding.Permit.S[0] ^= 1 },
	}
	for i, mutate := range mutations {
		r, e := DecodeSingle(data)
		if e != nil {
			t.Fatal(e)
		}
		mutate(r)
		encoded, e := Encode(r)
		if e != nil {
			t.Fatalf("mutation %d: %v", i, e)
		}
		hash, e := RequestHash(encoded)
		if e != nil || hash == original {
			t.Fatalf("mutation %d not bound (%v)", i, e)
		}
	}
	for _, mutate := range []func(*Request){func(r *Request) { r.Proof = []byte{1} }, func(r *Request) { r.Funding.Signature = []byte{2} }} {
		r, _ := DecodeSingle(data)
		mutate(r)
		encoded, e := Encode(r)
		if e != nil {
			t.Fatal(e)
		}
		hash, e := RequestHash(encoded)
		if e != nil || hash != original {
			t.Fatal("unsigned field changed hash")
		}
	}
}
func TestRejectNonzeroPaddingAndAlternateOffsets(t *testing.T) {
	v := fixture(t)
	data := unhex(t, v.Data)
	padding := append([]byte(nil), data...)
	at := bytes.Index(padding, []byte{0xde, 0xad, 0xbe, 0xef})
	if at < 0 {
		t.Fatal("missing proof")
	}
	padding[at+4] = 1
	if _, e := DecodeSingle(padding); e == nil {
		t.Fatal("nonzero padding accepted")
	}
	offset := append([]byte(nil), data...)
	offset[35] = 0
	if _, e := DecodeSingle(offset); e == nil {
		t.Fatal("bad offset accepted")
	}
}
func TestQuoteDomainAndSignatureBounds(t *testing.T) {
	v := fixture(t)
	d := domain(v)
	sig := unhex(t, v.Signature)
	d.Protocol[0] ^= 1
	if VerifyQuote(d, v.Quote, sig, common.HexToAddress(v.Trust.Signer)) == nil {
		t.Fatal("foreign protocol accepted")
	}
	d = domain(v)
	for _, bad := range []string{"-1", "not-a-number", new(big.Int).Lsh(big.NewInt(1), 256).String()} {
		q := v.Quote
		q.FeeAmount = bad
		if _, e := QuoteDigest(d, q); e == nil {
			t.Fatalf("invalid uint accepted %q", bad)
		}
	}
	for _, bad := range [][]byte{nil, sig[:63], append(sig, 0)} {
		if _, e := RecoverQuoteSigner(d, v.Quote, bad); e == nil {
			t.Fatal("invalid signature accepted")
		}
	}
	compact := append([]byte(nil), sig[:64]...)
	compact[32] |= (sig[64] - 27) << 7
	actual, e := RecoverQuoteSigner(d, v.Quote, compact)
	if e != nil || actual != common.HexToAddress(v.Trust.Signer) {
		t.Fatalf("compact signature: %v", e)
	}
}

func FuzzDecodeNeverPanics(f *testing.F) {
	b, e := os.ReadFile("testdata/sdk-vector.json")
	if e != nil {
		f.Fatal(e)
	}
	var v vector
	if e = json.Unmarshal(b, &v); e != nil {
		f.Fatal(e)
	}
	seed, e := hex.DecodeString(v.Data[2:])
	if e != nil {
		f.Fatal(e)
	}
	f.Add(seed)
	all, e := os.ReadFile("testdata/sdk-vectors.json")
	if e != nil {
		f.Fatal(e)
	}
	var vectors []vector
	if e = json.Unmarshal(all, &vectors); e != nil {
		f.Fatal(e)
	}
	for _, item := range vectors {
		f.Add(common.FromHex(item.Data))
	}
	f.Add([]byte{0, 1, 2})
	f.Fuzz(func(t *testing.T, b []byte) {
		r, e := DecodeSingle(b)
		if e != nil {
			return
		}
		encoded, e := Encode(r)
		if e != nil || !bytes.Equal(b, encoded) {
			t.Fatal("noncanonical decode")
		}
	})
}

func TestQuoteRejectsInvalidMixedChecksumAddress(t *testing.T) {
	v := fixture(t)
	q := v.Quote
	q.Token = "0x52908400098527886E0F7030069857D2E4169Ee7"
	if _, e := QuoteDigest(domain(v), q); e == nil {
		t.Fatal("ethers rejects invalid mixed address checksum")
	}
}

func TestOrdinaryOutputWithoutEnvelopeRemainsCanonicalAndBound(t *testing.T) {
	v := fixture(t)
	r, err := DecodeSingle(unhex(t, v.Data))
	if err != nil {
		t.Fatal(err)
	}
	original, err := Hash(r)
	if err != nil {
		t.Fatal(err)
	}
	r.Deliveries[0] = Delivery{}
	encoded, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSingle(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Deliveries) != 1 || len(decoded.Deliveries[0].Ciphertext) != 0 || len(decoded.Deliveries[0].Clue) != 0 {
		t.Fatal("lost explicit empty delivery slot")
	}
	hash, err := RequestHash(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if hash == original {
		t.Fatal("omitting envelope did not change bound intent")
	}
	roundtrip, err := Encode(decoded)
	if err != nil || !bytes.Equal(roundtrip, encoded) {
		t.Fatalf("empty delivery changed calldata: %v", err)
	}
}
