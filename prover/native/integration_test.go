package native

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

// Public deterministic test key only; never use this account with funds.
func creditCancellationRequest(t *testing.T) prover.ProvingRequest {
	t.Helper()
	wallet, err := core.DeriveWallet(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	context := core.ProtocolContext{ChainID: core.ElementFromUint64(31337), ProtocolAddress: core.ElementFromUint64(4660)}
	token, nonce, expiry, amount, shared := core.ElementFromUint64(22136), core.ElementFromUint64(123), core.ElementFromUint64(2000000000), core.ElementFromUint64(100), core.ElementFromUint64(456)
	keyHash, err := core.ComputeCreditSharedKeyHash(shared)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := core.ComputeCreditCommitment(context.ChainID, wallet.WalletIdentity, wallet.WalletIdentity, wallet.WalletIdentity, wallet.WalletIdentity, token, amount, nonce, expiry, keyHash)
	if err != nil {
		t.Fatal(err)
	}
	root := leaf
	path := make([]string, 16)
	for i := range path {
		path[i] = "0"
		root, err = core.Poseidon2Hash(root, core.ElementFromUint64(0))
		if err != nil {
			t.Fatal(err)
		}
	}
	nullifier, err := core.ComputeCreditNullifier(context.ChainID, shared, leaf)
	if err != nil {
		t.Fatal(err)
	}
	callContext := protocol.TransactionContext{ChainID: core.ElementToBigInt(context.ChainID), ProtocolAddress: common.BigToAddress(core.ElementToBigInt(context.ProtocolAddress)), TokenAddress: common.BigToAddress(core.ElementToBigInt(token))}
	params := protocol.CancelCreditParams{Token: callContext.TokenAddress, Owner: common.BigToAddress(core.ElementToBigInt(wallet.WalletIdentity)), Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: root.Bytes()}, Nullifier: nullifier.Bytes(), Amount: core.ElementToBigInt(amount), Deadline: big.NewInt(2000000000)}
	base, err := protocol.CancelCreditBaseInputs(callContext, params)
	if err != nil {
		t.Fatal(err)
	}
	words := make([]fr.Element, len(base))
	for i, value := range base {
		words[i] = core.BigIntToElement(new(big.Int).SetBytes(value[:]))
	}
	request := core.WalletAuthorizationRequest{Context: context, Operation: core.WalletCancelCredit, Capacity: 1, Words: words}
	signed, err := core.BuildWalletAuthorizationWitness(wallet, request)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := protocol.CancelCreditPublicInputs(callContext, params)
	if err != nil {
		t.Fatal(err)
	}
	public := make([]common.Hash, len(inputs))
	for i, value := range inputs {
		public[i] = value
	}
	proofRequest, err := prover.Request("credit_cancel", "evm", map[string]any{
		"public_key_x": signed.PublicKeyX, "public_key_y": signed.PublicKeyY, "signature": signed.Signature,
		"spender_payout": wallet.WalletIdentity.String(), "spender_wallet_identity": wallet.WalletIdentity.String(), "owner_wallet_identity": wallet.WalletIdentity.String(),
		"input_nonce": nonce.String(), "expiry": expiry.String(), "credit_shared_key": shared.String(), "leaf_index": "0", "merkle_path": path,
	}, public)
	if err != nil {
		t.Fatal(err)
	}
	return proofRequest
}

func TestRealOfficialWitnessExecutor(t *testing.T) {
	binary := os.Getenv("NOIR_EXECUTE_PATH")
	if binary == "" {
		t.Skip("set NOIR_EXECUTE_PATH to Noir 1.0.0-rc.6 noir-execute")
	}
	scratch := t.TempDir()
	t.Setenv("TMPDIR", scratch)
	request := creditCancellationRequest(t)
	artifact, _, err := assets(request.ProofStatement)
	if err != nil {
		t.Fatal(err)
	}
	executor := noirWitnessExecutor{binary: binary, timeout: time.Minute}
	witness, err := executor.Execute(context.Background(), artifact, request.Witness)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(witness)
	if !bytes.HasPrefix(witness, []byte{0x1f, 0x8b}) {
		t.Fatal("not a compressed ACVM witness")
	}
	var invalid *prover.LocalWitness
	err = request.Witness.WithInputs(func(inputs map[string]any) error {
		inputs["signature"] = [64]byte{}
		var err error
		invalid, err = prover.NewLocalWitness(inputs)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = executor.Execute(context.Background(), artifact, invalid); err == nil {
		t.Fatal("invalid signature accepted")
	}
	if strings.Contains(err.Error(), "public_key") || strings.Contains(err.Error(), "signature =") {
		t.Fatal("private diagnostics leaked")
	}
	leftovers, err := os.ReadDir(scratch)
	if err != nil || len(leftovers) != 0 {
		t.Fatal("temporary private witness files were retained", err)
	}

}

func TestRealNativeCreditCancellationProof(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	request := creditCancellationRequest(t)
	backend, err := NewNativeProver(NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
	if err != nil {
		t.Fatal(err)
	}
	phases := []string{}
	bundle, err := backend.Prove(context.Background(), request, prover.Options{OnProgress: func(event prover.Progress) { phases = append(phases, event.Phase) }})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(phases, ",") != "witness,proving,verifying,complete" {
		t.Fatal(phases)
	}
	verified, err := backend.Verify(context.Background(), bundle, request.ProofStatement, prover.Options{})
	if err != nil || !verified {
		t.Fatal("proof verification failed", err)
	}
	bundle.PublicInputs[0] = common.Hash{}
	if _, err = backend.Verify(context.Background(), bundle, request.ProofStatement, prover.Options{}); err == nil {
		t.Fatal("modified public inputs accepted")
	}
}
