package provenance

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ProofResult struct {
	Proof        string
	PublicInputs []string
}

// OfflineProver passes private openings only to an explicitly supplied local proving capability.
type OfflineProver interface {
	Prove(context.Context, ManifestIdentity, int, Witness, []string) (ProofResult, error)
}

func ExportSidecar(ctx context.Context, r WitnessRequest, certs []RootCertificate, t TrustContext, identity ManifestIdentity, prover OfflineProver, operationIndex uint32) (ProofSidecar, error) {
	if identity.SHA256 != t.ManifestSHA256 || len(certs) != len(r.Sources) {
		return ProofSidecar{}, errors.New("manifest or certificate count mismatch")
	}
	for _, p := range [][2]string{{r.ChainID, t.ChainID}, {r.ProtocolAddress, t.ProtocolAddress}, {r.ChainGeneration, t.ChainGeneration}} {
		if !equalNumber(p[0], p[1]) {
			return ProofSidecar{}, errors.New("witness deployment mismatch")
		}
	}
	for i, c := range certs {
		if e := AssertCertificateTrust(c, t); e != nil {
			return ProofSidecar{}, e
		}
		s := r.Sources[i]
		if !equalNumber(c.TreeID, s.TreeID) || !equalNumber(c.Root, s.Root) || !equalNumber(c.LeafCount, s.LeafCount) {
			return ProofSidecar{}, errors.New("certificate source mismatch")
		}
	}
	w, e := BuildWitness(r)
	if e != nil {
		return ProofSidecar{}, e
	}
	expected := make([]string, 11)
	for i, k := range publicInputNames {
		expected[i], _ = word(w[k].(string))
	}
	result, e := prover.Prove(ctx, identity, r.Capacity, w, expected)
	if e != nil {
		return ProofSidecar{}, e
	}
	if len(result.PublicInputs) != 11 {
		return ProofSidecar{}, errors.New("prover statement width")
	}
	for i, v := range result.PublicInputs {
		if !equalNumber(v, expected[i]) {
			return ProofSidecar{}, errors.New("prover statement mismatch")
		}
	}
	b, e := hex.DecodeString(strings.TrimPrefix(result.Proof, "0x"))
	if e != nil || !strings.HasPrefix(result.Proof, "0x") || len(b) == 0 || len(b) > 65536 {
		return ProofSidecar{}, errors.New("invalid proof bytes")
	}
	wireCertificates := make([]RootCertificate, len(certs))
	for i, c := range certs {
		normalized, e := certificateToWire(c)
		if e != nil {
			return ProofSidecar{}, e
		}
		wireCertificates[i] = normalized
	}
	return ProofSidecar{OperationIndex: strconv.FormatUint(uint64(operationIndex), 10), ProofBundle: ProofBundle{Format: "invisi-provenance-proof-v1", CircuitID: fmt.Sprintf("provenance_k%d", r.Capacity), ManifestSHA256: identity.SHA256, Target: "evm", PublicInputs: expected, Proof: result.Proof, RootCertificates: wireCertificates}}, nil
}
