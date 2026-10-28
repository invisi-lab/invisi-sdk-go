package receiver

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func makeVoucher(t *testing.T, recv common.Address) ipp.IPPVoucher {
	t.Helper()
	var mk, inv fr.Element // payer masterKey, random invoiceId
	mk.SetUint64(999)
	inv.SetUint64(7)
	nk, _ := core.DeriveNoteSecret(mk, inv)
	ic, _ := core.ComputeIdentityCommitment(nk, core.BigIntToElement(recv.Big()))
	var amount fr.Element
	amount.SetUint64(100_000_000)
	var nonce fr.Element
	nonce.SetUint64(12345)
	token := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	chainID := core.ElementFromUint64(84532)
	pc, _ := core.ComputeNoteCommitment(chainID, ic, core.BigIntToElement(token.Big()), amount, nonce)
	return ipp.IPPVoucher{
		V: 1, Network: "eip155:84532", Type: ipp.MessageTypeVoucher,
		Payload: ipp.IPPVoucherV1{
			ReceivingAddress: recv.Hex(),
			NoteCommitment:   core.ElementToHex(pc),
			TokenAddress:     token.Hex(),
			Operator:         "0x1234567890123456789012345678901234567890",
			Amount:           "100000000",
			Nonce:            core.ElementToHex(nonce),
			NoteSecret:       core.ElementToHex(nk),
			TreeID:           "1",
			Endpoint:         "https://pay.merchant.com",
		},
	}
}

func TestVerifyVoucher_OK(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	ok, err := VerifyVoucher(makeVoucher(t, recv), recv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected valid voucher")
	}
}

func TestVerifyVoucher_WrongAddress(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	other := common.HexToAddress("0x00000000000000000000000000000000bEEf0000")
	ok, _ := VerifyVoucher(makeVoucher(t, recv), other)
	if ok {
		t.Error("expected rejection for wrong receiving address")
	}
}

func TestVerifyVoucher_TamperedAmount(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	v := makeVoucher(t, recv)
	v.Payload.Amount = "999"
	ok, _ := VerifyVoucher(v, recv)
	if ok {
		t.Error("expected rejection for tampered amount")
	}
}

func TestVoucherToWithdrawRequest_NoSecret(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	v := makeVoucher(t, recv)
	req := VoucherToWithdrawRequest(v)
	if len(req.NoteCommitments) != 1 || req.NoteCommitments[0] != v.Payload.NoteCommitment {
		t.Error("noteCommitments mismatch")
	}

	// noteSecret and receivingAddress are non-custodial secrets: they stay
	// client-side for local proof generation and must never reach the wire.
	// Assert against the actual JSON.Marshal output, not just the struct shape,
	// so a future field addition to WithdrawRequest can't silently leak them.
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	wire := string(b)
	if strings.Contains(wire, "note_secret") || strings.Contains(wire, v.Payload.NoteSecret) {
		t.Errorf("withdraw request wire form leaks noteSecret: %s", wire)
	}
	if strings.Contains(wire, "receiving_address") || strings.Contains(wire, v.Payload.ReceivingAddress) {
		t.Errorf("withdraw request wire form leaks receivingAddress: %s", wire)
	}
}

// withdrawRequestFixtureVoucher returns a voucher whose envelope Network
// ("eip155:1") differs from every field value in the payload below. This
// catches an implementation that mistakenly sources the withdraw request's
// Network from the payload (which carries no network field of its own)
// instead of the envelope.
func withdrawRequestFixtureVoucher() ipp.IPPVoucher {
	return ipp.IPPVoucher{
		V:       1,
		Network: "eip155:1",
		Type:    ipp.MessageTypeVoucher,
		Payload: ipp.IPPVoucherV1{
			ReceivingAddress: "0xReceivingAddressPlaceholder",
			NoteCommitment:   "0xNoteCommitmentPlaceholder",
			TokenAddress:     "0xTokenAddressPlaceholder",
			Operator:         "0xOperatorPlaceholder",
			Amount:           "100000000",
			Nonce:            "0xNoncePlaceholder",
			NoteSecret:       "0xNoteSecretSecretPlaceholder",
			TreeID:           "42",
			Endpoint:         "https://pay.merchant.com",
		},
	}
}

func TestVoucherToWithdrawRequest_Fields(t *testing.T) {
	v := withdrawRequestFixtureVoucher()
	req := VoucherToWithdrawRequest(v)

	if req.Network != v.Network {
		t.Errorf("Network = %q, want envelope value %q", req.Network, v.Network)
	}
	if req.Operator != v.Payload.Operator {
		t.Errorf("Operator = %q, want %q", req.Operator, v.Payload.Operator)
	}
	if req.TokenAddress != v.Payload.TokenAddress {
		t.Errorf("TokenAddress = %q, want %q", req.TokenAddress, v.Payload.TokenAddress)
	}
	if req.TreeID != v.Payload.TreeID {
		t.Errorf("TreeID = %q, want %q", req.TreeID, v.Payload.TreeID)
	}
	if len(req.NoteCommitments) != 1 || req.NoteCommitments[0] != v.Payload.NoteCommitment {
		t.Errorf("NoteCommitments = %v, want [%q]", req.NoteCommitments, v.Payload.NoteCommitment)
	}
}

func TestVoucherToWithdrawRequest_WireShape(t *testing.T) {
	v := withdrawRequestFixtureVoucher()
	req := VoucherToWithdrawRequest(v)

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	want := []string{"network", "note_commitments", "operator", "token_address", "tree_id"}
	if len(keys) != len(want) {
		t.Fatalf("wire keys = %v, want exactly %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("wire keys = %v, want exactly %v", keys, want)
		}
	}
}

// TestVerifyVoucher_TamperedNetwork guards against a payer tampering with the
// voucher envelope's Network field while the NoteCommitment itself remains
// genuine (computed under the original, untampered chain ID). Because the
// authoritative chain ID is always re-derived from voucher.Network (not taken
// from a caller-supplied argument), recomputing NoteCommitment under the
// tampered network yields a mismatch and VerifyVoucher must reject it.
func TestVerifyVoucher_TamperedNetwork(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	v := makeVoucher(t, recv) // NoteCommitment computed under chainId 84532
	v.Network = "eip155:1"    // tamper: envelope now claims a different chain
	ok, _ := VerifyVoucher(v, recv)
	if ok {
		t.Error("expected rejection for voucher with tampered envelope network")
	}
}
