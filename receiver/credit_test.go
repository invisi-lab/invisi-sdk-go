package receiver

import (
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/internal/testvectors"
	"github.com/invisi-lab/invisi-sdk-go/payer"
)

func creditTestContext() payer.TransactionContext {
	return payer.TransactionContext{ChainID: big.NewInt(31337), ProtocolAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"), TokenAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"), Now: big.NewInt(800)}
}
func creditWord(n int64) [32]byte { return payer.Word(big.NewInt(n)) }
func creditSource() payer.InputSource {
	return payer.InputSource{TreeId: big.NewInt(1), MerkleRoot: creditWord(1)}
}
func creditEnvelope() []byte { b := bytes.Repeat([]byte{0x22}, 254); b[0] = 2; return b }
func drawTestParams(n int) (payer.CreditDrawParams, [][]byte) {
	p := payer.CreditDrawParams{Sources: []payer.InputSource{creditSource()}, Spender: common.HexToAddress("0x3333333333333333333333333333333333333333"), TotalAmount: big.NewInt(50), Deadline: big.NewInt(900)}
	for i := 0; i < n; i++ {
		p.Nullifiers = append(p.Nullifiers, creditWord(int64(i+2)))
		p.RemainderCommitments = append(p.RemainderCommitments, [32]byte{})
		p.RecoveryAmounts = append(p.RecoveryAmounts, [32]byte{})
		p.Expiries = append(p.Expiries, big.NewInt(1000))
	}
	return p, make([][]byte, 2*n)
}
func topTestParams() (payer.TopUpCreditParams, [][]byte, payer.FundingParams) {
	return payer.TopUpCreditParams{Creation: payer.CreateCreditParams{Owner: common.HexToAddress("0x3333333333333333333333333333333333333333"), CreditCommitment: creditWord(5), Expiry: big.NewInt(1000), Deadline: big.NewInt(900)}, CreditSource: creditSource(), CreditNullifier: creditWord(2), PreviousExpiry: big.NewInt(950)}, [][]byte{creditEnvelope(), creditEnvelope()}, payer.FundingParams{Payer: common.HexToAddress("0x4444444444444444444444444444444444444444"), Amount: big.NewInt(100), SigNonce: big.NewInt(7), Deadline: big.NewInt(1000), Signature: []byte{1}}
}

func TestCreditDrawProfilesAndPublicInputs(t *testing.T) {
	for _, v := range []struct{ count, size, width, bits int }{{1, 1, 16, 126}, {2, 12, 82, 130}, {12, 12, 82, 130}, {13, 50, 310, 132}, {50, 50, 310, 132}, {51, 120, 730, 133}, {120, 120, 730, 133}} {
		p, err := SelectCreditDrawCircuit(v.count)
		if err != nil || p.MaxSize != v.size || p.PublicInputCount != v.width || p.TotalBits != v.bits {
			t.Fatalf("profile %+v: %+v %v", v, p, err)
		}
		d, cs := drawTestParams(v.count)
		inputs, err := CreditDrawPublicInputs(creditTestContext(), d, cs)
		if err != nil || len(inputs) != v.width {
			t.Fatalf("inputs %d: %d %v", v.count, len(inputs), err)
		}
		if inputs[4] != creditWord(int64(v.count)) || inputs[6+2*v.size] != creditWord(2) {
			t.Fatal("input order")
		}
	}
	for _, n := range []int{-1, 0, 121} {
		if _, err := SelectCreditDrawCircuit(n); err == nil {
			t.Fatal("accepted invalid count", n)
		}
	}
}
func TestCreditDrawRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*payer.CreditDrawParams, *[][]byte)
	}{
		{"expired", func(p *payer.CreditDrawParams, _ *[][]byte) { p.Expiries[0] = big.NewInt(800) }},
		{"oversized expiry", func(p *payer.CreditDrawParams, _ *[][]byte) { p.Expiries[0] = new(big.Int).Lsh(big.NewInt(1), 64) }},
		{"total overflow", func(p *payer.CreditDrawParams, _ *[][]byte) { p.TotalAmount = new(big.Int).Lsh(big.NewInt(1), 126) }},
		{"recovery without remainder", func(p *payer.CreditDrawParams, _ *[][]byte) { p.RecoveryAmounts[0] = creditWord(1) }},
		{"noncanonical recovery", func(p *payer.CreditDrawParams, _ *[][]byte) { p.RecoveryAmounts[0] = payer.Word(fr.Modulus()) }},
		{"full draw ciphertext", func(_ *payer.CreditDrawParams, cs *[][]byte) { (*cs)[0] = creditEnvelope() }},
		{"partial draw envelope", func(p *payer.CreditDrawParams, cs *[][]byte) {
			p.RemainderCommitments[0] = creditWord(8)
			(*cs)[0] = creditEnvelope()
		}},
		{"duplicate remainder", func(p *payer.CreditDrawParams, cs *[][]byte) {
			p.RemainderCommitments[0] = creditWord(8)
			p.RemainderCommitments[1] = creditWord(8)
			for i := range *cs {
				(*cs)[i] = creditEnvelope()
			}
		}},
		{"array mismatch", func(p *payer.CreditDrawParams, _ *[][]byte) { p.RecoveryAmounts = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, cs := drawTestParams(1)
			if tt.name == "duplicate remainder" {
				p, cs = drawTestParams(2)
			}
			tt.mutate(&p, &cs)
			if _, err := ValidateCreditDrawRequest(creditTestContext(), p, cs); err == nil {
				t.Fatal("accepted invalid request")
			}
		})
	}
	p, cs := drawTestParams(1)
	p.RemainderCommitments[0] = creditWord(8)
	p.RecoveryAmounts[0] = creditWord(9)
	cs[0] = creditEnvelope()
	cs[1] = creditEnvelope()
	if _, err := BuildCreditDrawCalldata(BuildCreditDrawCalldataParams{Context: creditTestContext(), Params: p, Ciphertexts: cs, Proof: []byte{1, 2}}); err != nil {
		t.Fatal(err)
	}
}
func TestCancelCreditAuthorizationAndAnytime(t *testing.T) {
	c := creditTestContext()
	p := payer.CancelCreditParams{Source: creditSource(), Nullifier: creditWord(2), Owner: common.HexToAddress("0x3333333333333333333333333333333333333333"), Amount: big.NewInt(50), Deadline: big.NewInt(900)}
	in, err := CancelCreditPublicInputs(c, p)
	if err != nil || len(in) != 9 || in[7] != creditWord(50) {
		t.Fatal(in, err)
	}
	h, err := CancelCreditRequestHash(p)
	if err != nil || h.Hex() != "0xda727f5fd0d12608ee06abecdb122245f0d63e6f744f717042db7d99f457c666" {
		t.Fatal(h, err)
	}
	a := payer.CreditAuthorization{SigNonce: big.NewInt(4), Signature: []byte{1}}
	td, err := EncodeCancelCreditTypedData(c, p, a)
	if err != nil || td.PrimaryType != "CancelCredit" || td.Message["requestHash"] != h.Hex() {
		t.Fatal(td, err)
	}
	built, err := BuildCancelCreditCalldata(BuildCancelCreditCalldataParams{Context: c, Params: p, Proof: []byte{1, 2}, Authorization: a})
	if err != nil {
		t.Fatal(err)
	}
	checkCreditFraming(t, "cancelCredit", built.Data)
	a.Signature = nil
	if _, err := BuildCancelCreditCalldata(BuildCancelCreditCalldataParams{Context: c, Params: p, Proof: []byte{1}, Authorization: a}); err == nil {
		t.Fatal("missing authorization accepted")
	}
	p.Amount = new(big.Int).Lsh(big.NewInt(1), 126)
	if _, err := ValidateCancelCreditRequest(c, p); err == nil {
		t.Fatal("amount overflow accepted")
	}
}
func TestTopUpCreditSemanticsAndSigners(t *testing.T) {
	c := creditTestContext()
	p, cs, f := topTestParams()
	in, err := TopUpCreditPublicInputs(c, p, cs, f)
	if err != nil || len(in) != 48 || in[47] != creditWord(950) {
		t.Fatal(len(in), err)
	}
	// Hash locks the current nested tuple, ciphertext commitment and funding payer/amount.
	h, err := TopUpCreditRequestHash(p, cs, f)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := EncodeTopUpCreditTypedData(c, p, cs, payer.CreditAuthorization{SigNonce: big.NewInt(4)}, f)
	if err != nil {
		t.Fatal(err)
	}
	fund, err := EncodeTopUpCreditFundingTypedData(c, p, cs, f)
	if err != nil || owner.PrimaryType != "TopUpCredit" || fund.PrimaryType != "FundTransact" || owner.Message["requestHash"] != h.Hex() || fund.Message["requestHash"] != h.Hex() || owner.Message["owner"] == fund.Message["payer"] {
		t.Fatal(owner, fund, err)
	}
	built, err := BuildTopUpCreditCalldata(BuildTopUpCreditCalldataParams{Context: c, Params: p, Ciphertexts: cs, Funding: f, Proof: []byte{1, 2}, Authorization: payer.CreditAuthorization{SigNonce: big.NewInt(4), Signature: []byte{1}}})
	if err != nil || built.ApproveAmount.Cmp(big.NewInt(100)) != 0 {
		t.Fatal(built, err)
	}
	checkCreditFraming(t, "topUpCredit", built.Data)
	for _, expiry := range []*big.Int{big.NewInt(0), big.NewInt(800), big.NewInt(1001), new(big.Int).Lsh(big.NewInt(1), 64)} {
		p.PreviousExpiry = expiry
		if _, err := ValidateTopUpCreditRequest(c, p, cs, f); err == nil {
			t.Fatal("accepted previous expiry", expiry)
		}
	}
	p, cs, f = topTestParams()
	p.Creation.Nullifiers = [][32]byte{p.CreditNullifier}
	p.Creation.Sources = []payer.InputSource{creditSource()}
	if _, err := ValidateTopUpCreditRequest(c, p, cs, f); err == nil {
		t.Fatal("duplicate ordinary/Credit nullifier accepted")
	}
	p, cs, f = topTestParams()
	p.PreviousExpiry.Set(p.Creation.Expiry)
	if _, err := ValidateTopUpCreditRequest(c, p, cs, f); err != nil {
		t.Fatal("equal expiry rejected", err)
	}
	f.Signature = nil
	f.SigNonce = new(big.Int)
	f.Deadline = new(big.Int)
	other := common.HexToAddress("0x55")
	if _, err := BuildTopUpCreditCalldata(BuildTopUpCreditCalldataParams{Context: c, Params: p, Ciphertexts: cs, Funding: f, Proof: []byte{1}, Caller: &other, Authorization: payer.CreditAuthorization{SigNonce: big.NewInt(4), Signature: []byte{1}}}); err == nil {
		t.Fatal("direct funding by different caller accepted")
	}
}
func TestCreditNonceGuidesBoundaries(t *testing.T) {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	if _, err := CancelCreditNonceGuide(max); err == nil {
		t.Fatal("overflow accepted")
	}
	if _, err := CancelCreditNonceGuide(nil); err == nil {
		t.Fatal("nil nonce accepted")
	}
	guide, err := CancelCreditNonceGuide(big.NewInt(4))
	if err != nil || guide.AuthorizationNonce.Int64() != 4 || guide.OwnerNextNonce.Int64() != 5 {
		t.Fatal(guide, err)
	}
	addr := common.HexToAddress("0x33")
	g, err := TopUpCreditNonceGuide(payer.CreditNonceGuideParams{Owner: addr, Payer: addr, Caller: common.HexToAddress("0x44"), OwnerNonce: big.NewInt(4), PayerNonce: big.NewInt(4), FundingAmount: big.NewInt(100)})
	if err != nil || g.FundingNonce.Int64() != 5 || g.OwnerNextNonce.Int64() != 6 || g.PayerNextNonce.Int64() != 6 || !g.FundingSignatureRequired {
		t.Fatal(g, err)
	}
}
func checkCreditFraming(t *testing.T, name string, data []byte) {
	t.Helper()
	a, err := payer.ProtocolABI()
	if err != nil {
		t.Fatal(err)
	}
	m := a.Methods[name]
	if !bytes.Equal(data[:4], m.ID) {
		t.Fatal("selector")
	}
	args, err := m.Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	expected, err := payer.EncodeCreditProof([]byte{1, 2})
	if err != nil || !bytes.Equal(args[0].([]byte), expected) {
		t.Fatal("raw proof must be framed exactly once", err)
	}
}

// These vectors are produced independently by the protocol client, rather than by Go ABI packing.
func TestCreditLifecycleProtocolVectors(t *testing.T) {
	vectors := testvectors.Load(t)
	c := testvectors.Decode[payer.TransactionContext](t, vectors.Context)
	for _, name := range []string{"draw1", "draw2", "draw13", "draw51"} {
		t.Run(name, func(t *testing.T) {
			f := testvectors.Decode[struct {
				Params       payer.CreditDrawParams
				Ciphertexts  [][]byte
				PublicInputs [][32]byte
				Calldata     []byte
			}](t, vectors.Fixtures[name])
			got, err := CreditDrawPublicInputs(c, f.Params, f.Ciphertexts)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, f.PublicInputs) {
				t.Fatal("public inputs differ from protocol vector")
			}
			if len(f.Calldata) > 0 {
				b, err := BuildCreditDrawCalldata(BuildCreditDrawCalldataParams{Context: c, Params: f.Params, Ciphertexts: f.Ciphertexts, Proof: []byte{0x12, 0x34}})
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(b.Data, f.Calldata) {
					t.Fatal("calldata differs from protocol vector")
				}
			}
		})
	}
	auth := payer.CreditAuthorization{SigNonce: big.NewInt(4), Signature: []byte{0xab, 0xcd}}
	t.Run("cancel", func(t *testing.T) {
		f := testvectors.Decode[struct {
			Params       payer.CancelCreditParams
			PublicInputs [][32]byte
			RequestHash  common.Hash
			Calldata     []byte
		}](t, vectors.Fixtures["cancel"])
		got, err := CancelCreditPublicInputs(c, f.Params)
		if err != nil || !reflect.DeepEqual(got, f.PublicInputs) {
			t.Fatal("public inputs differ", err)
		}
		h, err := CancelCreditRequestHash(f.Params)
		if err != nil || h != f.RequestHash {
			t.Fatal("request hash differs", h, err)
		}
		td, err := EncodeCancelCreditTypedData(c, f.Params, auth)
		if err != nil {
			t.Fatal(err)
		}
		checkCreditTypedDataVector(t, vectors.Fixtures["cancel"], "typedData", td)
		b, err := BuildCancelCreditCalldata(BuildCancelCreditCalldataParams{Context: c, Params: f.Params, Proof: []byte{0x12, 0x34}, Authorization: auth})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b.Data, f.Calldata) {
			t.Fatal("calldata differs")
		}
	})
	t.Run("topUp", func(t *testing.T) {
		f := testvectors.Decode[struct {
			Params       payer.TopUpCreditParams
			Ciphertexts  [][]byte
			Funding      payer.FundingParams
			PublicInputs [][32]byte
			RequestHash  common.Hash
			Calldata     []byte
		}](t, vectors.Fixtures["topUp"])
		got, err := TopUpCreditPublicInputs(c, f.Params, f.Ciphertexts, f.Funding)
		if err != nil || !reflect.DeepEqual(got, f.PublicInputs) {
			t.Fatal("public inputs differ", err)
		}
		h, err := TopUpCreditRequestHash(f.Params, f.Ciphertexts, f.Funding)
		if err != nil || h != f.RequestHash {
			t.Fatal("request hash differs", h, err)
		}
		td, err := EncodeTopUpCreditTypedData(c, f.Params, f.Ciphertexts, auth, f.Funding)
		if err != nil {
			t.Fatal(err)
		}
		checkCreditTypedDataVector(t, vectors.Fixtures["topUp"], "typedData", td)
		td, err = EncodeTopUpCreditFundingTypedData(c, f.Params, f.Ciphertexts, f.Funding)
		if err != nil {
			t.Fatal(err)
		}
		checkCreditTypedDataVector(t, vectors.Fixtures["topUp"], "fundingTypedData", td)
		f.Funding.Signature = []byte{0xab, 0xcd}
		b, err := BuildTopUpCreditCalldata(BuildTopUpCreditCalldataParams{Context: c, Params: f.Params, Ciphertexts: f.Ciphertexts, Funding: f.Funding, Proof: []byte{0x12, 0x34}, Authorization: auth})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b.Data, f.Calldata) {
			t.Fatal("calldata differs")
		}
	})
}
func checkCreditTypedDataVector(t *testing.T, raw json.RawMessage, key string, got apitypes.TypedData) {
	t.Helper()
	var fixture map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	var want apitypes.TypedData
	if err := json.Unmarshal(fixture[key], &want); err != nil {
		t.Fatal(err)
	}
	want.Types["EIP712Domain"] = []apitypes.Type{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}
	expected, err := payer.TypedDataHash(want)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := payer.TypedDataHash(got)
	if err != nil || expected != actual {
		t.Fatal("typed data digest differs", actual, expected, err)
	}
}
