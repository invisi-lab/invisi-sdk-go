package prover

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestLocalWitnessCopiesInputsAndRedactsSerialization(t *testing.T) {
	source := map[string]any{"private": "sensitive-value"}
	w, e := NewLocalWitness(source)
	if e != nil {
		t.Fatal(e)
	}
	source["private"] = "modified"
	if e = w.WithInputs(func(inputs map[string]any) error {
		if inputs["private"] != "sensitive-value" {
			t.Fatal("aliased inputs")
		}
		inputs["private"] = "changed"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	raw, e := json.Marshal(w)
	if e != nil || strings.Contains(string(raw), "sensitive-value") {
		t.Fatal("witness serialization leaked")
	}
	if e = w.WithInputs(func(inputs map[string]any) error {
		if inputs["private"] != "sensitive-value" {
			t.Fatal("consumer mutated stored inputs")
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
}
func TestUnpublishedTargetAndWrongStatementWidthRejected(t *testing.T) {
	if _, e := Statement("credit_cancel", "noir-recursive", make([]common.Hash, 11)); e == nil {
		t.Fatal("unpublished proof target accepted")
	}
	if _, e := Statement("credit_cancel", "evm", make([]common.Hash, 12)); e == nil {
		t.Fatal("wrong statement width accepted")
	}
}

func TestProofBundleUsesTypeScriptExchangeFormat(t *testing.T) {
	original := ProofBundle{FormatVersion: 1, CircuitName: "credit_cancel", Target: "evm", BBVersion: "6.0.0", Proof: []byte{0x12, 0x34}, PublicInputs: []common.Hash{common.HexToHash("0x01")}}
	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"proofHex":"0x1234"`) || !strings.Contains(string(raw), `"formatVersion":1`) {
		t.Fatal(string(raw))
	}
	var decoded ProofBundle
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if common.Bytes2Hex(decoded.Proof) != "1234" || decoded.PublicInputs[0] != original.PublicInputs[0] {
		t.Fatal("proof exchange changed bytes")
	}
}

func TestProofBundleRejectsPreviousBBVersion(t *testing.T) {
	statement, err := Statement("credit_cancel", "evm", make([]common.Hash, 11))
	if err != nil {
		t.Fatal(err)
	}
	proof := make([]byte, 32)
	digest := sha256.Sum256(proof)
	bundle := ProofBundle{FormatVersion: 1, CircuitName: statement.CircuitName, Target: statement.Target, BBVersion: "6.0.0", ArtifactSHA256: statement.ArtifactSHA256, VKSHA256: statement.VKSHA256, Proof: proof, ProofSHA256: hex.EncodeToString(digest[:]), PublicInputs: statement.ExpectedPublicInputs}
	if err := AssertProofBundle(bundle, statement); err != nil {
		t.Fatal(err)
	}
	bundle.BBVersion = "5.2.1"
	if err := AssertProofBundle(bundle, statement); err == nil {
		t.Fatal("previous BB proof accepted")
	}
}

func TestRegistrationIsNotAProofCircuit(t *testing.T) {
	if _, err := Statement("wallet_registration", "evm", make([]common.Hash, 3)); err == nil {
		t.Fatal("removed registration circuit accepted")
	}
}
