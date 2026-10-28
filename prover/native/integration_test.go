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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

// Public deterministic test key only; never use this account with funds.
func registrationRequest(t *testing.T) prover.ProvingRequest {
	t.Helper()
	wallet, err := core.DeriveWallet(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	keysHash := crypto.Keccak256Hash(wallet.ViewingPublicKey, wallet.ClueKey)
	keyNumber := keysHash.Big()
	request := core.WalletAuthorizationRequest{
		Context:   core.ProtocolContext{ChainID: core.ElementFromUint64(31337), ProtocolAddress: core.ElementFromUint64(4660)},
		Operation: core.WalletRegisterAddress,
		Words:     []fr.Element{core.ElementFromUint64(31337), core.ElementFromUint64(4660), wallet.WalletIdentity, core.ElementFromUint64(0), core.ElementFromUint64(2000000000), core.BigIntToElement(new(big.Int).Rsh(new(big.Int).Set(keyNumber), 128)), core.BigIntToElement(new(big.Int).And(keyNumber, mask))},
	}
	signed, err := core.BuildWalletAuthorizationWitness(wallet, request)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := core.ComputeWalletAuthorizationDigest(request)
	if err != nil {
		t.Fatal(err)
	}
	high := new(big.Int).Rsh(digest.Big(), 128)
	low := new(big.Int).And(digest.Big(), mask)
	proofRequest, err := prover.Request("wallet_registration", "evm", map[string]any{"public_key_x": signed.PublicKeyX, "public_key_y": signed.PublicKeyY, "signature": signed.Signature}, []common.Hash{common.BigToHash(core.ElementToBigInt(wallet.WalletIdentity)), common.BigToHash(high), common.BigToHash(low)})
	if err != nil {
		t.Fatal(err)
	}
	return proofRequest
}

func TestRealOfficialWitnessExecutor(t *testing.T) {
	binary := os.Getenv("NOIR_EXECUTE_PATH")
	if binary == "" {
		t.Skip("set NOIR_EXECUTE_PATH to Noir 1.0.0-rc.4 noir-execute")
	}
	scratch := t.TempDir()
	t.Setenv("TMPDIR", scratch)
	request := registrationRequest(t)
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

func TestRealNativeRegistrationProof(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	request := registrationRequest(t)
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
