package receiver

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/tree"
)

func TestBuildCreditDrawWitness_SingleCredit(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	nk, err := core.DeriveNullifierKey(mk, invID)
	if err != nil {
		t.Fatal(err)
	}

	recvAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")

	// Build a tree with one credit commitment leaf
	mt, err := tree.NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	// Insert a dummy leaf (in real use, this would be a creditCommitment)
	var leaf fr.Element
	leaf.SetUint64(999)
	idx, err := mt.Insert(leaf)
	if err != nil {
		t.Fatal(err)
	}

	proof, err := mt.GetProof(idx)
	if err != nil {
		t.Fatal(err)
	}

	nkHash, err := core.ComputeNullifierKeyHash(nk)
	if err != nil {
		t.Fatal(err)
	}

	var payerComm fr.Element
	payerComm.SetUint64(42)

	witness, err := BuildCreditDrawWitness(BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{nk},
		ReceivingAddress:   recvAddr,
		TokenAddress:       common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Amounts:            []*big.Int{big.NewInt(50_000_000)},
		Nonces:             []fr.Element{{}},
		Expiries:           []*big.Int{big.NewInt(9999999999)},
		PayerCommitments:   []fr.Element{payerComm},
		NullifierKeyHashes: []fr.Element{nkHash},
		MerkleProofs:       []tree.MerkleProof{proof},
	})
	if err != nil {
		t.Fatal(err)
	}

	if witness.TotalAmount.Cmp(big.NewInt(50_000_000)) != 0 {
		t.Errorf("TotalAmount = %s, want 50000000", witness.TotalAmount)
	}
	if len(witness.NullifierKeys) != 1 {
		t.Errorf("NullifierKeys length = %d, want 1", len(witness.NullifierKeys))
	}
	if witness.Receiver != recvAddr {
		t.Error("Receiver mismatch")
	}
	if len(witness.RemainderCommitments) != 1 {
		t.Errorf("RemainderCommitments length = %d, want 1", len(witness.RemainderCommitments))
	}
	if len(witness.Expiries) != 1 {
		t.Errorf("Expiries length = %d, want 1", len(witness.Expiries))
	}
	if len(witness.PayerCommitments) != 1 {
		t.Errorf("PayerCommitments length = %d, want 1", len(witness.PayerCommitments))
	}
	if len(witness.NullifierKeyHashes) != 1 {
		t.Errorf("NullifierKeyHashes length = %d, want 1", len(witness.NullifierKeyHashes))
	}

	// TreeRoot should match the tree
	root := mt.Root()
	if !witness.TreeRoot.Equal(&root) {
		t.Errorf("TreeRoot mismatch")
	}
}

func TestBuildCreditDrawWitness_MultipleCredits(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	nk, err := core.DeriveNullifierKey(mk, invID)
	if err != nil {
		t.Fatal(err)
	}

	recvAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")

	mt, err := tree.NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	// Insert two dummy leaves
	var leaf1, leaf2 fr.Element
	leaf1.SetUint64(999)
	leaf2.SetUint64(1000)
	idx1, _ := mt.Insert(leaf1)
	idx2, _ := mt.Insert(leaf2)

	proof1, _ := mt.GetProof(idx1)
	proof2, _ := mt.GetProof(idx2)

	nkHash, _ := core.ComputeNullifierKeyHash(nk)

	var payerComm fr.Element
	payerComm.SetUint64(42)

	witness, err := BuildCreditDrawWitness(BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{nk, nk},
		ReceivingAddress:   recvAddr,
		TokenAddress:       common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Amounts:            []*big.Int{big.NewInt(50_000_000), big.NewInt(30_000_000)},
		Nonces:             []fr.Element{{}, {}},
		Expiries:           []*big.Int{big.NewInt(9999999999), big.NewInt(9999999999)},
		PayerCommitments:   []fr.Element{payerComm, payerComm},
		NullifierKeyHashes: []fr.Element{nkHash, nkHash},
		MerkleProofs:       []tree.MerkleProof{proof1, proof2},
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedTotal := big.NewInt(80_000_000)
	if witness.TotalAmount.Cmp(expectedTotal) != 0 {
		t.Errorf("TotalAmount = %s, want %s", witness.TotalAmount, expectedTotal)
	}
	if len(witness.NullifierKeys) != 2 {
		t.Errorf("NullifierKeys length = %d, want 2", len(witness.NullifierKeys))
	}
}

func TestBuildCreditDrawWitness_MismatchedLengths(t *testing.T) {
	_, err := BuildCreditDrawWitness(BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{{}, {}},
		ReceivingAddress:   common.Address{},
		TokenAddress:       common.Address{},
		Amounts:            []*big.Int{big.NewInt(100)},
		Nonces:             []fr.Element{{}},
		Expiries:           []*big.Int{big.NewInt(9999)},
		PayerCommitments:   []fr.Element{{}},
		NullifierKeyHashes: []fr.Element{{}},
		MerkleProofs:       []tree.MerkleProof{{}},
	})
	if err == nil {
		t.Error("expected error for mismatched lengths")
	}
}

func TestBuildCreditDrawWitness_EmptyInput(t *testing.T) {
	_, err := BuildCreditDrawWitness(BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{},
		ReceivingAddress:   common.Address{},
		TokenAddress:       common.Address{},
		Amounts:            []*big.Int{},
		Nonces:             []fr.Element{},
		Expiries:           []*big.Int{},
		PayerCommitments:   []fr.Element{},
		NullifierKeyHashes: []fr.Element{},
		MerkleProofs:       []tree.MerkleProof{},
	})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestBuildCreditDrawWitness_NilAmount(t *testing.T) {
	mt, err := tree.NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	var leaf fr.Element
	leaf.SetUint64(999)
	idx, _ := mt.Insert(leaf)
	proof, _ := mt.GetProof(idx)

	_, err = BuildCreditDrawWitness(BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{{}},
		ReceivingAddress:   common.Address{},
		TokenAddress:       common.Address{},
		Amounts:            []*big.Int{nil},
		Nonces:             []fr.Element{{}},
		Expiries:           []*big.Int{big.NewInt(9999)},
		PayerCommitments:   []fr.Element{{}},
		NullifierKeyHashes: []fr.Element{{}},
		MerkleProofs:       []tree.MerkleProof{proof},
	})
	if err == nil {
		t.Error("expected error for nil amount")
	}
}

func TestBuildCreditDrawWitness_InconsistentRoots(t *testing.T) {
	mt1, err := tree.NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	mt2, err := tree.NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf1, leaf2 fr.Element
	leaf1.SetUint64(999)
	leaf2.SetUint64(1000)
	idx1, _ := mt1.Insert(leaf1)
	idx2, _ := mt2.Insert(leaf2)

	proof1, _ := mt1.GetProof(idx1)
	proof2, _ := mt2.GetProof(idx2)

	_, err = BuildCreditDrawWitness(BuildCreditDrawWitnessParams{
		NullifierKeys:      []fr.Element{{}, {}},
		ReceivingAddress:   common.Address{},
		TokenAddress:       common.Address{},
		Amounts:            []*big.Int{big.NewInt(100), big.NewInt(200)},
		Nonces:             []fr.Element{{}, {}},
		Expiries:           []*big.Int{big.NewInt(9999), big.NewInt(9999)},
		PayerCommitments:   []fr.Element{{}, {}},
		NullifierKeyHashes: []fr.Element{{}, {}},
		MerkleProofs:       []tree.MerkleProof{proof1, proof2},
	})
	if err == nil {
		t.Error("expected error for inconsistent merkle roots")
	}
}

func TestBuildCreditDrawCalldata_EncodesCorrectly(t *testing.T) {
	result, err := BuildCreditDrawCalldata(CreditDrawParams{
		Nullifiers:           []common.Hash{common.HexToHash("0xababababababababababababababababababababababababababababababababab")},
		TotalAmount:          big.NewInt(100_000_000),
		ReceivingAddress:     common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:           common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:                common.Hex2Bytes("deadbeef"),
		TreeID:               big.NewInt(1),
		RemainderCommitments: []common.Hash{},
		Expiry:               big.NewInt(9999999999),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
	if len(result.Nullifiers) != 1 {
		t.Errorf("Nullifiers length = %d, want 1", len(result.Nullifiers))
	}
	if result.TotalAmount.Cmp(big.NewInt(100_000_000)) != 0 {
		t.Errorf("TotalAmount = %s, want 100000000", result.TotalAmount)
	}
}

func TestBuildCreditDrawCalldata_WithRemainderCommitments(t *testing.T) {
	result, err := BuildCreditDrawCalldata(CreditDrawParams{
		Nullifiers:       []common.Hash{common.HexToHash("0xababababababababababababababababababababababababababababababababab")},
		TotalAmount:      big.NewInt(50_000_000),
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:       common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:            common.Hex2Bytes("deadbeef"),
		TreeID:           big.NewInt(1),
		RemainderCommitments: []common.Hash{
			common.HexToHash("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		},
		Expiry: big.NewInt(9999999999),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestBuildCreditDrawCalldata_EmptyNullifiers(t *testing.T) {
	_, err := BuildCreditDrawCalldata(CreditDrawParams{
		Nullifiers:           []common.Hash{},
		TotalAmount:          big.NewInt(0),
		ReceivingAddress:     common.Address{},
		MerkleRoot:           common.Hash{},
		Proof:                []byte{},
		TreeID:               big.NewInt(0),
		RemainderCommitments: []common.Hash{},
		Expiry:               big.NewInt(0),
	})
	if err == nil {
		t.Error("expected error for empty nullifiers")
	}
}

func TestBuildCreditDrawCalldata_MultipleNullifiers(t *testing.T) {
	nullifiers := []common.Hash{
		common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
	}
	result, err := BuildCreditDrawCalldata(CreditDrawParams{
		Nullifiers:           nullifiers,
		TotalAmount:          big.NewInt(200_000_000),
		ReceivingAddress:     common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:           common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:                common.Hex2Bytes("deadbeef"),
		TreeID:               big.NewInt(1),
		RemainderCommitments: []common.Hash{},
		Expiry:               big.NewInt(9999999999),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Nullifiers) != 2 {
		t.Errorf("Nullifiers length = %d, want 2", len(result.Nullifiers))
	}
	if result.TotalAmount.Cmp(big.NewInt(200_000_000)) != 0 {
		t.Errorf("TotalAmount = %s, want 200000000", result.TotalAmount)
	}
}
