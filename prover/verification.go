package prover

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func AssertProofBundle(bundle ProofBundle, statement ProofStatement) error {
	expected, e := Statement(statement.CircuitName, statement.Target, statement.ExpectedPublicInputs)
	if e != nil {
		return e
	}
	if statement.ArtifactSHA256 != expected.ArtifactSHA256 || statement.VKSHA256 != expected.VKSHA256 || bundle.FormatVersion != 1 || bundle.BBVersion != "6.0.0" || bundle.CircuitName != statement.CircuitName || bundle.Target != statement.Target || bundle.ArtifactSHA256 != statement.ArtifactSHA256 || bundle.VKSHA256 != statement.VKSHA256 || len(bundle.Proof) == 0 || len(bundle.PublicInputs) != len(statement.ExpectedPublicInputs) {
		return errors.New("proof bundle identity mismatch")
	}
	if len(bundle.Proof)%32 != 0 || (bundle.Target == "noir-recursive" && len(bundle.Proof) != 458*32) {
		return errors.New("invalid proof field encoding or recursive width")
	}
	h := sha256.Sum256(bundle.Proof)
	if hex.EncodeToString(h[:]) != bundle.ProofSHA256 {
		return errors.New("proof digest mismatch")
	}
	for i, v := range bundle.PublicInputs {
		if v != statement.ExpectedPublicInputs[i] {
			return errors.New("proof public inputs mismatch")
		}
	}
	return nil
}
