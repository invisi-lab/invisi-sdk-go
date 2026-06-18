package receiver

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/tree"
)

// buildTestMerkleTree creates a tree with N payments and returns
// the tree, the nonces used, and the leaf indices.
func buildTestMerkleTree(t *testing.T, nk fr.Element, recvAddr common.Address, tokenAddr common.Address, chainID *big.Int, amounts []*big.Int) (*tree.InMemoryMerkleTree, []fr.Element, []int) {
	t.Helper()
	mt, err := tree.NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	recvElem := core.BigIntToElement(new(big.Int).SetBytes(recvAddr.Bytes()))
	ic, err := core.ComputeIdentityCommitment(nk, recvElem)
	if err != nil {
		t.Fatal(err)
	}

	tokenElem := core.BigIntToElement(new(big.Int).SetBytes(tokenAddr.Bytes()))
	chainElem := core.BigIntToElement(chainID)

	nonces := make([]fr.Element, len(amounts))
	indices := make([]int, len(amounts))
	for i, amt := range amounts {
		nonce, err := core.GenerateNonce()
		if err != nil {
			t.Fatal(err)
		}
		nonces[i] = nonce
		amountElem := core.BigIntToElement(amt)

		commitment, err := core.ComputePaymentCommitment(chainElem, ic, tokenElem, amountElem, nonce)
		if err != nil {
			t.Fatal(err)
		}

		idx, err := mt.Insert(commitment)
		if err != nil {
			t.Fatal(err)
		}
		indices[i] = idx
	}
	return mt, nonces, indices
}

func TestBuildWithdrawWitness_SinglePayment(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	nk, err := core.DeriveNullifierKey(mk, invID)
	if err != nil {
		t.Fatal(err)
	}

	recvAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")
	tokenAddr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	chainID := big.NewInt(84532)
	amounts := []*big.Int{big.NewInt(100_000_000)}

	mt, nonces, indices := buildTestMerkleTree(t, nk, recvAddr, tokenAddr, chainID, amounts)

	proofs := make([]tree.MerkleProof, len(indices))
	for i, idx := range indices {
		proof, err := mt.GetProof(idx)
		if err != nil {
			t.Fatal(err)
		}
		proofs[i] = proof
	}

	witness, err := BuildWithdrawWitness(BuildWithdrawWitnessParams{
		NullifierKeys:    []fr.Element{nk},
		ReceivingAddress: recvAddr,
		TokenAddress:     tokenAddr,
		Amounts:          amounts,
		Nonces:           nonces,
		MerkleProofs:     proofs,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify TotalAmount
	if witness.TotalAmount.Cmp(big.NewInt(100_000_000)) != 0 {
		t.Errorf("TotalAmount = %s, want 100000000", witness.TotalAmount)
	}

	// Verify lengths
	if len(witness.NullifierKeys) != 1 {
		t.Errorf("NullifierKeys length = %d, want 1", len(witness.NullifierKeys))
	}
	if len(witness.Amounts) != 1 {
		t.Errorf("Amounts length = %d, want 1", len(witness.Amounts))
	}
	if len(witness.Nonces) != 1 {
		t.Errorf("Nonces length = %d, want 1", len(witness.Nonces))
	}
	if len(witness.MerkleProofs) != 1 {
		t.Errorf("MerkleProofs length = %d, want 1", len(witness.MerkleProofs))
	}

	// TreeRoot should match the tree
	root := mt.Root()
	if !witness.TreeRoot.Equal(&root) {
		t.Errorf("TreeRoot mismatch")
	}

	// Receiver should be set
	if witness.Receiver != recvAddr {
		t.Errorf("Receiver = %s, want %s", witness.Receiver.Hex(), recvAddr.Hex())
	}
}

func TestBuildWithdrawWitness_MultiplePayments(t *testing.T) {
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	nk, err := core.DeriveNullifierKey(mk, invID)
	if err != nil {
		t.Fatal(err)
	}

	recvAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")
	tokenAddr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	chainID := big.NewInt(84532)
	amounts := []*big.Int{big.NewInt(100_000_000), big.NewInt(200_000_000), big.NewInt(50_000_000)}

	mt, nonces, indices := buildTestMerkleTree(t, nk, recvAddr, tokenAddr, chainID, amounts)

	proofs := make([]tree.MerkleProof, len(indices))
	for i, idx := range indices {
		proof, err := mt.GetProof(idx)
		if err != nil {
			t.Fatal(err)
		}
		proofs[i] = proof
	}

	// All payments use the same nullifierKey
	nks := make([]fr.Element, len(amounts))
	for i := range nks {
		nks[i] = nk
	}

	witness, err := BuildWithdrawWitness(BuildWithdrawWitnessParams{
		NullifierKeys:    nks,
		ReceivingAddress: recvAddr,
		TokenAddress:     tokenAddr,
		Amounts:          amounts,
		Nonces:           nonces,
		MerkleProofs:     proofs,
	})
	if err != nil {
		t.Fatal(err)
	}

	// TotalAmount = 100 + 200 + 50 = 350
	expectedTotal := big.NewInt(350_000_000)
	if witness.TotalAmount.Cmp(expectedTotal) != 0 {
		t.Errorf("TotalAmount = %s, want %s", witness.TotalAmount, expectedTotal)
	}

	if len(witness.NullifierKeys) != 3 {
		t.Errorf("NullifierKeys length = %d, want 3", len(witness.NullifierKeys))
	}
}

func TestBuildWithdrawWitness_MismatchedLengths(t *testing.T) {
	_, err := BuildWithdrawWitness(BuildWithdrawWitnessParams{
		NullifierKeys:    []fr.Element{{}, {}},
		ReceivingAddress: common.Address{},
		TokenAddress:     common.Address{},
		Amounts:          []*big.Int{big.NewInt(100)}, // length mismatch
		Nonces:           []fr.Element{{}},
		MerkleProofs:     []tree.MerkleProof{{}},
	})
	if err == nil {
		t.Error("expected error for mismatched lengths")
	}
}

func TestBuildWithdrawWitness_EmptyInput(t *testing.T) {
	_, err := BuildWithdrawWitness(BuildWithdrawWitnessParams{
		NullifierKeys:    []fr.Element{},
		ReceivingAddress: common.Address{},
		TokenAddress:     common.Address{},
		Amounts:          []*big.Int{},
		Nonces:           []fr.Element{},
		MerkleProofs:     []tree.MerkleProof{},
	})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestBuildWithdrawWitness_ConsistentRoots(t *testing.T) {
	// All merkle proofs must share the same root
	var mk fr.Element
	mk.SetUint64(12345)
	var invID fr.Element
	invID.SetUint64(1)

	nk, _ := core.DeriveNullifierKey(mk, invID)
	recvAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")
	tokenAddr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	chainID := big.NewInt(84532)

	// Build two separate trees
	mt1, nonces1, indices1 := buildTestMerkleTree(t, nk, recvAddr, tokenAddr, chainID, []*big.Int{big.NewInt(100)})
	mt2, nonces2, indices2 := buildTestMerkleTree(t, nk, recvAddr, tokenAddr, chainID, []*big.Int{big.NewInt(200)})

	proof1, _ := mt1.GetProof(indices1[0])
	proof2, _ := mt2.GetProof(indices2[0])

	_, err := BuildWithdrawWitness(BuildWithdrawWitnessParams{
		NullifierKeys:    []fr.Element{nk, nk},
		ReceivingAddress: recvAddr,
		TokenAddress:     tokenAddr,
		Amounts:          []*big.Int{big.NewInt(100), big.NewInt(200)},
		Nonces:           append(nonces1, nonces2...),
		MerkleProofs:     []tree.MerkleProof{proof1, proof2},
	})
	if err == nil {
		t.Error("expected error for inconsistent merkle roots")
	}
}
