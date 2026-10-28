package payer_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/internal/testvectors"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
	p "github.com/invisi-lab/invisi-sdk-go/payer"
	"math/big"
	"testing"
)

type fixture[T any] struct {
	Params                                                     T
	Outputs                                                    []p.ShieldOutput
	Ciphertexts                                                [][]byte
	Funding                                                    p.FundingParams
	PublicInputs, UserPublicInputs                             [][32]byte
	Digest, RequestHash, FundingRequestHash, MemberRequestHash common.Hash
	Calldata                                                   []byte
}

func load[T any](t *testing.T, name string) (p.TransactionContext, fixture[T]) {
	t.Helper()
	v := testvectors.Load(t)
	return testvectors.Decode[p.TransactionContext](t, v.Context), testvectors.Decode[fixture[T]](t, v.Fixtures[name])
}
func require(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func equalInputs(t *testing.T, got, want [][32]byte) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("width %d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("public input%d got %x want%x", i, got[i], want[i])
		}
	}
}
func expectedTypedHash(t *testing.T, name, key string) common.Hash {
	t.Helper()
	v := testvectors.Load(t)
	var fields map[string]json.RawMessage
	require(t, json.Unmarshal(v.Fixtures[name], &fields))
	var td apitypes.TypedData
	require(t, json.Unmarshal(fields[key], &td))
	td.Types["EIP712Domain"] = []apitypes.Type{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}
	h, err := p.TypedDataHash(td)
	require(t, err)
	return h
}
func TestCurrentTransactStatementVectors(t *testing.T) {
	for _, count := range []int{1, 15, 57, 113} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			c, f := load[p.CrossTreeTransactParams](t, fmt.Sprintf("transact%d", count))
			business, err := p.TransactPublicInputs(c, f.Params, f.Ciphertexts, p.FundingParams{})
			require(t, err)
			equalInputs(t, business, f.PublicInputs)
			full, err := p.TransactUserPublicInputs(c, f.Params, f.Ciphertexts, p.FundingParams{})
			require(t, err)
			equalInputs(t, full, f.UserPublicInputs)
			td, err := p.EncodeTransactTypedData(c, f.Params, f.Ciphertexts, p.FundingParams{})
			require(t, err)
			hash, err := p.TypedDataHash(td)
			require(t, err)
			if hash != f.Digest {
				t.Fatalf("typed authorization got %s want%s", hash, f.Digest)
			}
		})
	}
}
func TestCurrentDirectCalldataAndFundingHash(t *testing.T) {
	for _, name := range []string{"single", "transact1", "crossTwo"} {
		t.Run(name, func(t *testing.T) {
			c, f := load[p.CrossTreeTransactParams](t, name)
			r, err := p.BuildTransactCalldata(p.BuildTransactCalldataParams{Context: c, Proof: []byte{0x12, 0x34}, Params: f.Params, Ciphertexts: f.Ciphertexts})
			require(t, err)
			if !bytes.Equal(r.Data, f.Calldata) {
				t.Fatal("current calldata differs from protocol ABI vector")
			}
			if f.FundingRequestHash != (common.Hash{}) {
				hash, err := p.TransactFundingRequestHash([]byte{0x12, 0x34}, f.Params, f.Ciphertexts)
				require(t, err)
				if hash != f.FundingRequestHash {
					t.Fatal("funding request hash mismatch")
				}
			}
		})
	}
	c, f := load[p.CrossTreeTransactParams](t, "transact1")
	hash, err := p.AggregateMemberRequestHash(c, p.TransactRequest{Params: f.Params, Ciphertexts: f.Ciphertexts})
	require(t, err)
	if hash != f.MemberRequestHash {
		t.Fatal("proof-independent aggregate member hash mismatch")
	}
}
func TestCurrentShieldVector(t *testing.T) {
	c, f := load[p.ShieldParams](t, "shield")
	params := p.BuildShieldCalldataParams{Context: c, Params: f.Params, Outputs: f.Outputs, Ciphertexts: f.Ciphertexts, Funding: f.Funding}
	result, err := p.BuildShieldCalldata(params)
	require(t, err)
	if !bytes.Equal(result.Data, f.Calldata) {
		t.Fatal("shield calldata mismatch")
	}
	hash, err := p.ShieldFundingRequestHash(f.Params, f.Outputs, f.Ciphertexts)
	require(t, err)
	if hash != f.RequestHash {
		t.Fatal("shield funding request hash mismatch")
	}
	td, err := p.EncodeShieldFundingTypedData(params)
	require(t, err)
	shieldDigest, err := p.TypedDataHash(td)
	require(t, err)
	if shieldDigest != expectedTypedHash(t, "shield", "fundingTypedData") {
		t.Fatal("Shield funding authorization mismatch")
	}
	if td.PrimaryType != "FundShield" {
		t.Fatal("incorrect shield funding type")
	}
	params.Funding.Amount = big.NewInt(101)
	if _, err := p.BuildShieldCalldata(params); err == nil {
		t.Fatal("accepted funding/output imbalance")
	}
}
func TestCurrentCreditCreationAndAuthorizationVectors(t *testing.T) {
	c, f := load[p.CreateCreditParams](t, "create")
	a := p.CreditAuthorization{SigNonce: big.NewInt(4), Signature: []byte{0xab, 0xcd}}
	inputs, err := p.CreateCreditPublicInputs(c, f.Params, f.Ciphertexts, f.Funding)
	require(t, err)
	equalInputs(t, inputs, f.PublicInputs)
	hash, err := p.CreateCreditRequestHash(f.Params, f.Ciphertexts, f.Funding)
	require(t, err)
	if hash != f.RequestHash {
		t.Fatal("creation request hash mismatch")
	}
	td, err := p.EncodeCreateCreditTypedData(c, f.Params, f.Ciphertexts, a, f.Funding)
	require(t, err)
	digest, err := p.TypedDataHash(td)
	require(t, err)
	if digest != expectedTypedHash(t, "create", "typedData") {
		t.Fatal("Credit owner authorization mismatch")
	}
	fundingTD, err := p.EncodeCreateCreditFundingTypedData(c, f.Params, f.Ciphertexts, f.Funding)
	require(t, err)
	fundingDigest, err := p.TypedDataHash(fundingTD)
	require(t, err)
	if fundingDigest != expectedTypedHash(t, "create", "fundingTypedData") {
		t.Fatal("Credit funding authorization mismatch")
	}
	if fundingDigest == digest {
		t.Fatal("owner and funding signatures must be distinct")
	}
	f.Funding.Signature = []byte{0xab, 0xcd}
	result, err := p.BuildCreateCreditCalldata(p.BuildCreateCreditCalldataParams{Context: c, Proof: []byte{0x12, 0x34}, Params: f.Params, Ciphertexts: f.Ciphertexts, Funding: f.Funding, Authorization: a})
	require(t, err)
	if !bytes.Equal(result.Data, f.Calldata) {
		t.Fatal("credit calldata / direct proof framing mismatch")
	}
}
func TestCurrentBoundaryValidation(t *testing.T) {
	for _, shape := range [][2]int{{0, 1}, {0, 0}, {-1, 2}, {1, 168}, {169, 0}} {
		if _, err := p.SelectTransactCapacity(shape[0], shape[1]); err == nil {
			t.Fatal("retired/invalid shape accepted", shape)
		}
	}
	c, f := load[p.CrossTreeTransactParams](t, "single")
	tests := map[string]func(*p.CrossTreeTransactParams){
		"zero tree":           func(r *p.CrossTreeTransactParams) { r.Sources[0].TreeId = big.NewInt(0) },
		"zero nullifier":      func(r *p.CrossTreeTransactParams) { r.Nullifiers[0] = [32]byte{} },
		"field overflow":      func(r *p.CrossTreeTransactParams) { r.Nullifiers[0] = p.Word(core.FieldModulus()) },
		"negative withdrawal": func(r *p.CrossTreeTransactParams) { r.WithdrawalAmount = big.NewInt(-1) },
		"deadline overflow":   func(r *p.CrossTreeTransactParams) { r.Deadline = new(big.Int).Lsh(big.NewInt(1), 64) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			_, v := load[p.CrossTreeTransactParams](t, "single")
			mutate(&v.Params)
			if _, err := p.BuildTransactCalldata(p.BuildTransactCalldataParams{Context: c, Proof: []byte{1}, Params: v.Params, Ciphertexts: v.Ciphertexts}); err == nil {
				t.Fatal("accepted invalid current request")
			}
		})
	}
	f.Params.Sources = nil
	f.Params.Nullifiers = nil
	f.Params.OutputCommitments = append(f.Params.OutputCommitments, p.Word(big.NewInt(9)))
	f.Ciphertexts = append(f.Ciphertexts, f.Ciphertexts[0])
	funding := p.FundingParams{Payer: c.ProtocolAddress, Amount: big.NewInt(100)}
	if _, err := p.BuildTransactCalldata(p.BuildTransactCalldataParams{Context: c, Proof: []byte{1}, Params: f.Params, Ciphertexts: f.Ciphertexts, Funding: funding}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.BuildTransactCalldata(p.BuildTransactCalldataParams{Context: c, Proof: []byte{1}, Params: f.Params, Ciphertexts: f.Ciphertexts}); err == nil {
		t.Fatal("accepted unfunded zero inputs")
	}
}
func TestFundingAndNonceOrder(t *testing.T) {
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	caller := common.HexToAddress("0x2222222222222222222222222222222222222222")
	r, err := p.CreateCreditNonceGuide(p.CreditNonceGuideParams{Owner: owner, Payer: owner, Caller: caller, OwnerNonce: big.NewInt(7), PayerNonce: big.NewInt(7), FundingAmount: big.NewInt(1)})
	require(t, err)
	if r.AuthorizationNonce.Int64() != 7 || r.FundingNonce.Int64() != 8 || r.OwnerNextNonce.Int64() != 9 || r.PayerNextNonce.Int64() != 9 || !r.FundingSignatureRequired {
		t.Fatal("wrong shared signer nonce order")
	}
	for _, bad := range []p.FundingParams{{Amount: big.NewInt(1)}, {Payer: owner}, {SigNonce: big.NewInt(1)}, {Permit: p.TokenPermit{V: 27}}} {
		if _, err := p.NormalizeFunding(bad); err == nil {
			t.Fatal("accepted noncanonical funding")
		}
	}
	f, err := p.NormalizeFunding(p.FundingParams{Payer: owner, Amount: big.NewInt(1)})
	require(t, err)
	if err := p.ValidateFundingSubmission(f, &caller, nil); err == nil {
		t.Fatal("accepted unsigned funding by other caller")
	}
	f.Signature = []byte{1}
	f.Deadline = big.NewInt(1)
	if err := p.ValidateFundingSubmission(f, &caller, big.NewInt(2)); err == nil {
		t.Fatal("accepted expired funding signature")
	}
}
func TestAggregateCurrentRequests(t *testing.T) {
	c, f := load[p.CrossTreeTransactParams](t, "single")
	requests := make([]p.TransactRequest, 4)
	for i := range requests {
		params := f.Params
		params.Nullifiers = [][32]byte{p.Word(big.NewInt(int64(100 + i)))}
		params.OutputCommitments = [][32]byte{p.Word(big.NewInt(int64(200 + i)))}
		requests[i] = p.TransactRequest{Params: params, Ciphertexts: f.Ciphertexts}
	}
	result, err := p.BuildTransactAggregateCalldata(p.BuildTransactAggregateCalldataParams{Context: c, Proof: []byte{0x12, 0x34}, Requests: requests})
	require(t, err)
	if result.Capacity != 14 || result.ApproveAmount.Sign() != 0 {
		t.Fatal("wrong aggregate result")
	}
	a, err := p.ProtocolABI()
	require(t, err)
	method, err := a.MethodById(result.Data[:4])
	require(t, err)
	if method.RawName != "transact" || method.Inputs[1].Type.T != abi.SliceTy {
		t.Fatal("wrong aggregate method")
	}
	business, err := p.AggregatePublicInputs(c, requests)
	require(t, err)
	if len(business) != 56*4 {
		t.Fatal("wrong router statement width")
	}
	requests[1].Params.Nullifiers = requests[0].Params.Nullifiers
	if _, err := p.BuildTransactAggregateCalldata(p.BuildTransactAggregateCalldataParams{Context: c, Proof: []byte{1}, Requests: requests}); err == nil {
		t.Fatal("accepted duplicate aggregate nullifier")
	}
}
func TestBuildShieldFromRequestAndCreditKeyVerification(t *testing.T) {
	c, f := load[p.ShieldParams](t, "shield")
	amount := "100"
	endpoint := "https://pay.example.com"
	request := ipp.IPPRequest{V: 1, Type: ipp.MessageTypeRequest, Network: "eip155:31337", Payload: ipp.IPPPayloadV1{IdentityCommitment: common.Hash(f.Outputs[0].IdentityCommitment).Hex(), TokenAddress: c.TokenAddress.Hex(), Operator: c.ProtocolAddress.Hex(), PaymentTypes: []ipp.PaymentType{ipp.PaymentTypeOpen}, Mode: "single", Amount: &amount, Endpoint: &endpoint}}
	params := p.BuildShieldFromRequestParams{Request: request, Nonce: f.Outputs[0].Nonce, NoteCommitment: f.Params.OutputCommitments[0], Amount: big.NewInt(100), ProtocolAddress: c.ProtocolAddress, Deadline: f.Params.Deadline, Ciphertext: f.Ciphertexts[0], Funding: f.Funding}
	result, err := p.BuildShieldFromRequest(params)
	require(t, err)
	if !bytes.Equal(result.Data, f.Calldata) {
		t.Fatal("IPP shield calldata mismatch")
	}
	params.Request.Payload.PaymentTypes = []ipp.PaymentType{ipp.PaymentTypeBlind}
	if _, err := p.BuildShieldFromRequest(params); err == nil {
		t.Fatal("accepted shield for blind-only request")
	}
	key := core.ElementFromUint64(75)
	hash, err := core.ComputeCreditSharedKeyHash(key)
	require(t, err)
	v := p.VerifyCreditCommitmentParams{ChainID: core.ElementFromUint64(1), SpenderIdentityCommitment: core.ElementFromUint64(2), OwnerIdentityCommitment: core.ElementFromUint64(3), TokenAddress: core.ElementFromUint64(4), Amount: core.ElementFromUint64(100), Nonce: core.ElementFromUint64(42), Expiry: core.ElementFromUint64(2000000000), CreditSharedKey: key}
	v.CreditCommitment, err = core.ComputeCreditCommitment(v.ChainID, v.SpenderIdentityCommitment, v.OwnerIdentityCommitment, v.TokenAddress, v.Amount, v.Nonce, v.Expiry, hash)
	require(t, err)
	valid, err := p.VerifyCreditCommitment(v)
	require(t, err)
	if !valid {
		t.Fatal("valid shared key rejected")
	}
	v.CreditSharedKey = core.ElementFromUint64(76)
	valid, err = p.VerifyCreditCommitment(v)
	require(t, err)
	if valid {
		t.Fatal("accepted substituted shared key")
	}
}

func TestRawRequestHashesRejectMissingIntegers(t *testing.T) {
	for _, run := range []func() error{
		func() error {
			_, err := p.TransactFundingRequestHash(nil, p.CrossTreeTransactParams{}, nil)
			return err
		},
		func() error { _, err := p.ShieldFundingRequestHash(p.ShieldParams{}, nil, nil); return err },
		func() error {
			_, err := p.ShieldFundingRequestHash(p.ShieldParams{Deadline: big.NewInt(1)}, []p.ShieldOutput{{}}, nil)
			return err
		},
	} {
		if err := run(); err == nil {
			t.Fatal("missing ABI integer accepted")
		}
	}
}
func TestAggregateProtocolCalldataGolden(t *testing.T) {
	v := testvectors.Load(t)
	c := testvectors.Decode[p.TransactionContext](t, v.Context)
	f := testvectors.Decode[struct {
		Requests     []p.TransactRequest
		PublicInputs [][32]byte
		Calldata     []byte
	}](t, v.Fixtures["aggregate4"])
	result, err := p.BuildTransactAggregateCalldata(p.BuildTransactAggregateCalldataParams{Context: c, Proof: []byte{0x12, 0x34}, Requests: f.Requests})
	require(t, err)
	if !bytes.Equal(result.Data, f.Calldata) {
		t.Fatal("aggregate overload calldata differs from independent protocol vector")
	}
	inputs, err := p.AggregatePublicInputs(c, f.Requests)
	require(t, err)
	equalInputs(t, inputs, f.PublicInputs)
	for _, count := range []int{8, 12} {
		requests := make([]p.TransactRequest, count)
		for i := range requests {
			request := f.Requests[i%4]
			request.Params.Nullifiers = [][32]byte{p.Word(big.NewInt(int64(100 + i)))}
			request.Params.OutputCommitments = [][32]byte{p.Word(big.NewInt(int64(200 + i)))}
			requests[i] = request
		}
		if _, err := p.BuildTransactAggregateCalldata(p.BuildTransactAggregateCalldataParams{Context: c, Proof: []byte{0x12, 0x34}, Requests: requests}); err != nil {
			t.Fatal(err)
		}
	}
}
