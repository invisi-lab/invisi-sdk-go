package provenance

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type ManifestEntry struct {
	Capacity       int    `json:"capacity"`
	Artifact       string `json:"artifact"`
	VK             string `json:"vk"`
	ArtifactSHA256 string `json:"artifactSha256"`
	VKSHA256       string `json:"vkSha256"`
}
type Manifest struct {
	Family       string          `json:"family"`
	Target       string          `json:"target"`
	NoirVersion  string          `json:"noirVersion"`
	BBVersion    string          `json:"bbVersion"`
	Depth        int             `json:"depth"`
	PublicInputs []string        `json:"publicInputs"`
	Circuits     []ManifestEntry `json:"circuits"`
}
type ManifestIdentity struct {
	Manifest Manifest
	SHA256   string
}

func AuthenticateManifest(data []byte, expected string) (ManifestIdentity, error) {
	h := sha256.Sum256(data)
	if len(expected) != 64 || hex.EncodeToString(h[:]) != expected {
		return ManifestIdentity{}, errors.New("manifest mismatch")
	}
	var m Manifest
	if e := json.Unmarshal(data, &m); e != nil {
		return ManifestIdentity{}, e
	}
	if m.Family != "provenance" || m.Target != "evm" || m.NoirVersion != "1.0.0" || m.BBVersion != "6.0.0" || m.Depth != 32 || len(m.PublicInputs) != 11 || len(m.Circuits) != 7 {
		return ManifestIdentity{}, errors.New("unsupported provenance manifest")
	}
	for i, k := range publicInputNames {
		if m.PublicInputs[i] != k {
			return ManifestIdentity{}, errors.New("public ABI mismatch")
		}
	}
	seen := map[int]bool{}
	for _, c := range m.Circuits {
		if !SupportedCapacity(c.Capacity) || seen[c.Capacity] || len(c.ArtifactSHA256) != 64 || len(c.VKSHA256) != 64 {
			return ManifestIdentity{}, errors.New("invalid manifest circuits")
		}
		seen[c.Capacity] = true
	}
	return ManifestIdentity{m, expected}, nil
}

type ProofBundle struct {
	Format           string            `json:"format"`
	CircuitID        string            `json:"circuitId"`
	ManifestSHA256   string            `json:"manifestSha256"`
	Target           string            `json:"target"`
	PublicInputs     []string          `json:"publicInputs"`
	Proof            string            `json:"proof"`
	RootCertificates []RootCertificate `json:"rootCertificates"`
}
type ProofSidecar struct {
	OperationIndex string      `json:"operationIndex"`
	ProofBundle    ProofBundle `json:"proofBundle"`
}
type TrustContext struct {
	Scope
	ChainGeneration    string
	ProviderID         string
	SigningKeys        map[string]ed25519.PublicKey
	SigningKeyValidity map[string]KeyValidity
	ManifestSHA256     string
}

func AssertCertificateTrust(c RootCertificate, t TrustContext) error {
	for _, p := range [][2]string{{c.ChainID, t.ChainID}, {c.ProtocolAddress, t.ProtocolAddress}, {c.ChainGeneration, t.ChainGeneration}, {c.ProviderID, t.ProviderID}} {
		if !equalNumber(p[0], p[1]) {
			return errors.New("certificate deployment mismatch")
		}
	}
	id, e := word(c.SigningKeyID)
	if e != nil {
		return e
	}
	if interval, ok := t.SigningKeyValidity[id]; ok {
		issued, e := Uint(c.IssuedAt, 64)
		if e != nil {
			return e
		}
		start, e := rotationTime(interval.ActivatedAt)
		if e != nil {
			return e
		}
		if issued.Uint64() < start {
			return errors.New("certificate before key activation")
		}
		if interval.RetiredAt != "" {
			end, e := rotationTime(interval.RetiredAt)
			if e != nil {
				return e
			}
			if issued.Uint64() >= end {
				return errors.New("certificate after key retirement")
			}
		}
	}
	if !VerifyRootCertificate(c, t.SigningKeys[id]) {
		return errors.New("untrusted certificate signature")
	}
	return nil
}

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
	return ProofSidecar{strconv.FormatUint(uint64(operationIndex), 10), ProofBundle{"invisi-provenance-proof-v1", fmt.Sprintf("provenance_k%d", r.Capacity), identity.SHA256, "evm", expected, result.Proof, wireCertificates}}, nil
}

func certificateToWire(c RootCertificate) (RootCertificate, error) {
	if _, e := RootCertificatePayload(c); e != nil {
		return RootCertificate{}, e
	}
	data, e := json.Marshal(c)
	if e != nil {
		return RootCertificate{}, e
	}
	var values map[string]string
	if e := json.Unmarshal(data, &values); e != nil {
		return RootCertificate{}, e
	}
	for k, v := range values {
		if k == "signature" {
			continue
		}
		n, e := Uint(v, 256)
		if e != nil {
			return RootCertificate{}, e
		}
		switch k {
		case "providerId", "signingKeyId", "root", "confirmedBlockHash", "decisionLogHash", "previousCheckpointHash":
			values[k] = fmt.Sprintf("0x%064x", n)
		case "protocolAddress":
			values[k] = fmt.Sprintf("0x%040x", n)
		default:
			values[k] = n.String()
		}
	}
	data, e = json.Marshal(values)
	if e != nil {
		return RootCertificate{}, e
	}
	var result RootCertificate
	e = json.Unmarshal(data, &result)
	return result, e
}

func ValidateProofBundle(b ProofBundle, t TrustContext) error {
	if b.Format != "invisi-provenance-proof-v1" || b.Target != "evm" || b.ManifestSHA256 != t.ManifestSHA256 || !strings.HasPrefix(b.Proof, "0x") {
		return errors.New("unsupported proof bundle")
	}
	proof, e := hex.DecodeString(strings.TrimPrefix(b.Proof, "0x"))
	if e != nil || len(proof) == 0 || len(proof) > 65536 {
		return errors.New("invalid proof bytes")
	}
	if !strings.HasPrefix(b.CircuitID, "provenance_k") {
		return errors.New("unsupported provenance circuit")
	}
	capacity, e := strconv.Atoi(strings.TrimPrefix(b.CircuitID, "provenance_k"))
	if e != nil || !SupportedCapacity(capacity) || b.CircuitID != fmt.Sprintf("provenance_k%d", capacity) || len(b.PublicInputs) != 11 {
		return errors.New("unsupported provenance circuit or ABI")
	}
	p := make([]*big.Int, 11)
	for i, v := range b.PublicInputs {
		if len(v) != 66 || !strings.HasPrefix(v, "0x") {
			return errors.New("public inputs must be canonical words")
		}
		p[i], e = Field(v)
		if e != nil {
			return e
		}
	}
	for _, pair := range [][2]string{{p[0].String(), t.ChainID}, {p[1].String(), t.ProtocolAddress}, {p[3].String(), t.ChainGeneration}} {
		if !equalNumber(pair[0], pair[1]) {
			return errors.New("proof deployment mismatch")
		}
	}
	if p[4].Cmp(big.NewInt(int64(len(b.RootCertificates)))) != 0 || p[6].Sign() == 0 || p[6].Cmp(big.NewInt(int64(capacity))) > 0 || p[4].Sign() == 0 || p[4].Cmp(p[6]) > 0 {
		return errors.New("proof shape mismatch")
	}
	if p[7].BitLen() > 128 || p[8].BitLen() > 128 {
		return errors.New("operation digest limbs")
	}
	sources := []Source{}
	for _, cert := range b.RootCertificates {
		if e := AssertCertificateTrust(cert, t); e != nil {
			return e
		}
		sources = append(sources, Source{cert.TreeID, cert.Root, cert.LeafCount})
	}
	sd, e := SourcesDigest(t.ChainGeneration, sources)
	if e != nil {
		return e
	}
	if sd != p[5].String() {
		return errors.New("proof sources digest mismatch")
	}
	scope, e := ScopeHash(t.Scope)
	if e != nil {
		return e
	}
	operation := new(big.Int).Or(new(big.Int).Lsh(new(big.Int).Set(p[7]), 128), p[8])
	binding, e := BindingDigest(scope, p[2].String(), t.ChainGeneration, p[4].String(), sd, p[6].String(), operation.String(), p[9].String())
	if e != nil {
		return e
	}
	if binding != p[10].String() {
		return errors.New("proof binding mismatch")
	}
	return nil
}
func assertPublicSubmission(value any) error {
	data, e := json.Marshal(value)
	if e != nil {
		return e
	}
	var decoded any
	if e := json.Unmarshal(data, &decoded); e != nil {
		return e
	}
	var visit func(any) error
	visit = func(v any) error {
		switch node := v.(type) {
		case map[string]any:
			for k, item := range node {
				switch strings.ToLower(k) {
				case "opening", "openings", "nonce", "nonces", "privatekey", "private_key", "ipn", "witness", "owners", "amounts", "paths", "artifacturl", "vkurl", "verificationkey":
					return errors.New("private or executable material is not allowed in provenance submissions")
				}
				if e := visit(item); e != nil {
					return e
				}
			}
		case []any:
			for _, item := range node {
				if e := visit(item); e != nil {
					return e
				}
			}
		}
		return nil
	}
	return visit(decoded)
}

func ProofBundleToWire(bundle ProofBundle) (ProofBundle, error) {
	result := bundle
	result.PublicInputs = make([]string, len(bundle.PublicInputs))
	for i, v := range bundle.PublicInputs {
		n, e := Field(v)
		if e != nil {
			return ProofBundle{}, e
		}
		result.PublicInputs[i], _ = word(n.String())
	}
	result.RootCertificates = make([]RootCertificate, len(bundle.RootCertificates))
	for i, c := range bundle.RootCertificates {
		normalized, e := certificateToWire(c)
		if e != nil {
			return ProofBundle{}, e
		}
		result.RootCertificates[i] = normalized
	}
	return result, nil
}

func ProofSidecarsToWire(sidecars []ProofSidecar) ([]ProofSidecar, error) {
	result := make([]ProofSidecar, len(sidecars))
	for i, sidecar := range sidecars {
		index, e := Uint(sidecar.OperationIndex, 32)
		if e != nil {
			return nil, e
		}
		bundle, e := ProofBundleToWire(sidecar.ProofBundle)
		if e != nil {
			return nil, e
		}
		result[i] = ProofSidecar{OperationIndex: index.String(), ProofBundle: bundle}
	}
	return result, nil
}
