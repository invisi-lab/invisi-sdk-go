package receiver

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
	"github.com/invisi-lab/invisi-sdk-go/tree"
)

func TestInvoiceFields(t *testing.T) {
	inv := Invoice{
		InvoiceID:          fr.Element{},
		NullifierKey:       fr.Element{},
		IdentityCommitment: common.Hash{},
		ReceivingAddress:   common.Address{},
		TokenAddress:       common.Address{},
		Amount:             big.NewInt(100),
		TreeID:             big.NewInt(1),
		Operator:           common.Address{},
	}
	if inv.Amount.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Amount = %s, want 100", inv.Amount)
	}
}

func TestGenerateInvoiceParamsFields(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	p := GenerateInvoiceParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: common.HexToAddress("0xdead"),
		TokenAddress:     common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Amount:           big.NewInt(100_000_000),
		TreeID:           big.NewInt(1),
		Operator:         common.HexToAddress("0x1234"),
	}
	if p.Amount.Cmp(big.NewInt(100_000_000)) != 0 {
		t.Errorf("Amount = %s, want 100000000", p.Amount)
	}
}

func TestBatchCircuitInfoFields(t *testing.T) {
	info := BatchCircuitInfo{Name: "batch_settle_1", MaxSize: 1}
	if info.Name != "batch_settle_1" {
		t.Errorf("Name = %s, want batch_settle_1", info.Name)
	}
	if info.MaxSize != 1 {
		t.Errorf("MaxSize = %d, want 1", info.MaxSize)
	}
}

func TestProposeParamsAndResult(t *testing.T) {
	p := ProposeParams{
		IdentityCommitment: common.Hash{},
		TokenAddress:       common.Address{},
		Amount:             big.NewInt(100),
		ChainID:            big.NewInt(84532),
	}
	if p.ChainID.Cmp(big.NewInt(84532)) != 0 {
		t.Errorf("ChainID = %s, want 84532", p.ChainID)
	}

	r := ProposeResult{
		Nonce:             common.Hash{},
		PaymentCommitment: common.Hash{},
	}
	_ = r
}

func TestCreatePaymentRequestParamsFields(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(42)
	var invID fr.Element
	invID.SetUint64(1)
	endpoint := "https://pay.example.com"

	p := CreatePaymentRequestParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: common.HexToAddress("0xdead"),
		TokenAddress:     common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		ChainID:          big.NewInt(84532),
		Operator:         common.HexToAddress("0x1234"),
		PaymentTypes:     []ipp.PaymentType{ipp.PaymentTypeOpen, ipp.PaymentTypeBlind},
		Mode:             "single",
		Amount:           big.NewInt(100_000_000),
		Endpoint:         &endpoint,
	}
	if len(p.PaymentTypes) != 2 {
		t.Errorf("PaymentTypes length = %d, want 2", len(p.PaymentTypes))
	}
}

func TestWithdrawWitnessFields(t *testing.T) {
	w := WithdrawWitness{
		NullifierKeys: []fr.Element{{}},
		Amounts:       []*big.Int{big.NewInt(100)},
		Nonces:        []fr.Element{{}},
		MerkleProofs:  []tree.MerkleProof{{}},
		TreeRoot:      fr.Element{},
		TotalAmount:   big.NewInt(100),
		Receiver:      common.Address{},
	}
	if w.TotalAmount.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("TotalAmount = %s, want 100", w.TotalAmount)
	}
}

func TestWithdrawCalldataFields(t *testing.T) {
	c := WithdrawCalldata{
		Data:        []byte{0xab},
		Nullifiers:  []common.Hash{{}},
		TotalAmount: big.NewInt(100),
	}
	if len(c.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(c.Data))
	}
}

func TestBuildWithdrawWitnessParamsFields(t *testing.T) {
	p := BuildWithdrawWitnessParams{
		NullifierKeys:    []fr.Element{{}},
		ReceivingAddress: common.Address{},
		TokenAddress:     common.Address{},
		Amounts:          []*big.Int{big.NewInt(100)},
		Nonces:           []fr.Element{{}},
		MerkleProofs:     []tree.MerkleProof{{}},
	}
	if len(p.NullifierKeys) != 1 {
		t.Errorf("NullifierKeys length = %d, want 1", len(p.NullifierKeys))
	}
}

func TestWithdrawParamsFields(t *testing.T) {
	p := WithdrawParams{
		Nullifiers:       []common.Hash{{}},
		TotalAmount:      big.NewInt(100),
		ReceivingAddress: common.Address{},
		MerkleRoot:       common.Hash{},
		Proof:            []byte{0xde, 0xad},
		TreeID:           big.NewInt(1),
	}
	if len(p.Proof) != 2 {
		t.Errorf("Proof length = %d, want 2", len(p.Proof))
	}
}

func TestCreditDrawWitnessFields(t *testing.T) {
	w := CreditDrawWitness{
		NullifierKeys:        []fr.Element{{}},
		Amounts:              []*big.Int{big.NewInt(50)},
		Nonces:               []fr.Element{{}},
		MerkleProofs:         []tree.MerkleProof{{}},
		TreeRoot:             fr.Element{},
		TotalAmount:          big.NewInt(50),
		Receiver:             common.Address{},
		RemainderCommitments: []fr.Element{{}},
		Expiries:             []*big.Int{big.NewInt(9999999999)},
		PayerCommitments:     []fr.Element{{}},
		NullifierKeyHashes:   []fr.Element{{}},
	}
	if w.TotalAmount.Cmp(big.NewInt(50)) != 0 {
		t.Errorf("TotalAmount = %s, want 50", w.TotalAmount)
	}
}

func TestCreditDrawCalldataFields(t *testing.T) {
	c := CreditDrawCalldata{
		Data:        []byte{0xab},
		Nullifiers:  []common.Hash{{}},
		TotalAmount: big.NewInt(100),
	}
	if len(c.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(c.Data))
	}
}

func TestBuildCreditDrawWitnessParamsFields(t *testing.T) {
	p := BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{{}},
		ReceivingAddress:   common.Address{},
		TokenAddress:       common.Address{},
		Amounts:            []*big.Int{big.NewInt(50)},
		Nonces:             []fr.Element{{}},
		Expiries:           []*big.Int{big.NewInt(9999999999)},
		PayerCommitments:   []fr.Element{{}},
		NullifierKeyHashes: []fr.Element{{}},
		MerkleProofs:       []tree.MerkleProof{{}},
	}
	if len(p.NullifierKeys) != 1 {
		t.Errorf("NullifierKeys length = %d, want 1", len(p.NullifierKeys))
	}
}

func TestCreditDrawParamsFields(t *testing.T) {
	p := CreditDrawParams{
		Nullifiers:           []common.Hash{{}},
		TotalAmount:          big.NewInt(100),
		ReceivingAddress:     common.Address{},
		MerkleRoot:           common.Hash{},
		Proof:                []byte{0xde, 0xad},
		TreeID:               big.NewInt(1),
		RemainderCommitments: []common.Hash{{}},
		Expiry:               big.NewInt(9999999999),
	}
	if len(p.Proof) != 2 {
		t.Errorf("Proof length = %d, want 2", len(p.Proof))
	}
}

func TestCreditReclaimParamsFields(t *testing.T) {
	p := CreditReclaimParams{
		Nullifier:        common.Hash{},
		Amount:           big.NewInt(100),
		ReceivingAddress: common.Address{},
		MerkleRoot:       common.Hash{},
		Proof:            []byte{0xde, 0xad},
		TreeID:           big.NewInt(1),
		Expiry:           big.NewInt(9999999999),
	}
	if len(p.Proof) != 2 {
		t.Errorf("Proof length = %d, want 2", len(p.Proof))
	}
}

func TestCreditReclaimCalldataFields(t *testing.T) {
	c := CreditReclaimCalldata{
		Data: []byte{0xab},
	}
	if len(c.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(c.Data))
	}
}

func TestRefundPayOpenParamsFields(t *testing.T) {
	p := RefundPayOpenParams{
		IdentityCommitment: common.Hash{},
		RefundAmount:       big.NewInt(500),
		Nonce:              common.Hash{},
		TreeID:             big.NewInt(1),
		RefundLink:         common.Hash{},
	}
	if p.RefundAmount.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("RefundAmount = %s, want 500", p.RefundAmount)
	}
}

func TestRefundPayOpenCalldataFields(t *testing.T) {
	c := RefundPayOpenCalldata{
		Data: []byte{0xab},
	}
	if len(c.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(c.Data))
	}
}
