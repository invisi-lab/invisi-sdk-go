package sdk

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

// GasReference is a synchronous pre-proof reference, not a transaction gas limit.
// Use the application's geth transaction pipeline for final estimation and sending.
type GasReference struct {
	GasUnits uint64 `json:"gasUnits"`
}

type GasBenchmarkShape struct {
	Kind             string `json:"kind"`
	Target           string `json:"target"`
	Capacity         int    `json:"capacity"`
	NoteInputs       int    `json:"noteInputs"`
	NoteOutputs      int    `json:"noteOutputs"`
	SourceTrees      int    `json:"sourceTrees"`
	CreditInputs     int    `json:"creditInputs"`
	CreditOutputs    int    `json:"creditOutputs"`
	CreditRemainders int    `json:"creditRemainders"`
	Funding          string `json:"funding"`
	Permit           bool   `json:"permit"`
	DeliveryBytes    int    `json:"deliveryBytes"`
}
type GasBenchmarkDeployment struct {
	ChainID         string `json:"chainId"`
	ProtocolAddress string `json:"protocolAddress"`
	RuntimeCodeHash string `json:"runtimeCodeHash"`
	TokenAddress    string `json:"tokenAddress"`
}
type GasBenchmarkCircuit struct {
	CircuitName    string `json:"circuitName"`
	Target         string `json:"target"`
	ArtifactSHA256 string `json:"artifactSha256"`
	VKSHA256       string `json:"vkSha256"`
}
type GasBenchmarkSample struct {
	GasUsed         string `json:"gasUsed"`
	TransactionHash string `json:"transactionHash"`
	BlockNumber     string `json:"blockNumber"`
	BlockHash       string `json:"blockHash"`
	CalldataBytes   int    `json:"calldataBytes"`
}
type GasBenchmarkFingerprints struct {
	ProtocolSourceSHA256    string `json:"protocolSourceSha256"`
	DirtyDiffSHA256         string `json:"dirtyDiffSha256"`
	CompiledArtifactsSHA256 string `json:"compiledArtifactsSha256"`
	CreditManifestSHA256    string `json:"creditManifestSha256"`
	UserManifestSHA256      string `json:"userManifestSha256"`
	SDKSourceSHA256         string `json:"sdkSourceSha256"`
}
type GasBenchmarkToolchain struct {
	Solidity      string `json:"solidity"`
	Barretenberg  string `json:"barretenberg"`
	Noir          string `json:"noir"`
	EVMVersion    string `json:"evmVersion"`
	OptimizerRuns int    `json:"optimizerRuns"`
	Anvil         string `json:"anvil"`
}
type GasBenchmarkRuntimeContract struct {
	Address         string `json:"address"`
	RuntimeCodeHash string `json:"runtimeCodeHash"`
}
type GasBenchmarkSource struct {
	ProtocolCommit   string                                 `json:"protocolCommit"`
	Fingerprints     *GasBenchmarkFingerprints              `json:"fingerprints,omitempty"`
	Toolchain        *GasBenchmarkToolchain                 `json:"toolchain,omitempty"`
	BlockGasLimit    string                                 `json:"blockGasLimit,omitempty"`
	RuntimeContracts map[string]GasBenchmarkRuntimeContract `json:"runtimeContracts,omitempty"`
	MeasuredAt       string                                 `json:"measuredAt"`
	Chain            string                                 `json:"chain"`
	Hardfork         string                                 `json:"hardfork"`
	Token            string                                 `json:"token"`
	Verifier         string                                 `json:"verifier"`
	Assumptions      []string                               `json:"assumptions"`
}
type GasBenchmarkMeasurement struct {
	ID      string               `json:"id"`
	Shape   GasBenchmarkShape    `json:"shape"`
	Circuit *GasBenchmarkCircuit `json:"circuit"`
	Samples []GasBenchmarkSample `json:"samples"`
}
type GasBenchmarkProfile struct {
	FormatVersion int                       `json:"formatVersion"`
	ID            string                    `json:"id"`
	Deployment    GasBenchmarkDeployment    `json:"deployment"`
	Source        GasBenchmarkSource        `json:"source"`
	Measurements  []GasBenchmarkMeasurement `json:"measurements"`
}

//go:embed data/gas-benchmarks.json
var gasBenchmarkJSON []byte

var defaultGasProfiles = func() []GasBenchmarkProfile {
	profiles, err := DefineGasBenchmarks(gasBenchmarkJSON)
	if err != nil {
		panic(err)
	}
	return profiles
}()

var gasWord = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)
var gasDigest = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
var gasCommit = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
var gasDecimal = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

// DefineGasBenchmarks parses trusted JSON receipt records in the same format as TS.
// Validation checks their format, not their publisher's honesty.
func DefineGasBenchmarks(data []byte) ([]GasBenchmarkProfile, error) {
	var profiles []GasBenchmarkProfile
	if err := json.Unmarshal(data, &profiles); err != nil || profiles == nil {
		return nil, errors.New("invalid gas benchmark catalog")
	}
	if err := validateGasProfiles(profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}
func validateGasProfiles(profiles []GasBenchmarkProfile) error {
	for _, p := range profiles {
		if p.FormatVersion != 1 || p.ID == "" || !gasDecimal.MatchString(p.Deployment.ChainID) ||
			!common.IsHexAddress(p.Deployment.ProtocolAddress) || !common.IsHexAddress(p.Deployment.TokenAddress) ||
			!gasWord.MatchString(p.Deployment.RuntimeCodeHash) || p.Source.Verifier != "real" ||
			!gasCommit.MatchString(p.Source.ProtocolCommit) || p.Source.MeasuredAt == "" ||
			p.Source.Assumptions == nil || len(p.Measurements) == 0 {
			return errors.New("invalid gas benchmark profile")
		}
		if f := p.Source.Fingerprints; f != nil {
			for _, fingerprint := range []string{f.ProtocolSourceSHA256, f.DirtyDiffSHA256, f.CompiledArtifactsSHA256, f.CreditManifestSHA256, f.UserManifestSHA256, f.SDKSourceSHA256} {
				if !gasDigest.MatchString(fingerprint) {
					return errors.New("invalid gas benchmark source fingerprint")
				}
			}
		}
		if c := p.Source.Toolchain; c != nil && (c.Solidity == "" || c.Barretenberg == "" || c.Noir == "" || c.EVMVersion == "" || c.OptimizerRuns < 0 || c.Anvil == "") {
			return errors.New("invalid gas benchmark toolchain")
		}
		if p.Source.BlockGasLimit != "" {
			limit, valid := new(big.Int).SetString(p.Source.BlockGasLimit, 10)
			if !valid || !gasDecimal.MatchString(p.Source.BlockGasLimit) || limit.Sign() <= 0 {
				return errors.New("invalid gas benchmark block gas limit")
			}
		}
		for name, contract := range p.Source.RuntimeContracts {
			if name == "" || !common.IsHexAddress(contract.Address) || !gasWord.MatchString(contract.RuntimeCodeHash) {
				return errors.New("invalid gas benchmark runtime contract")
			}
		}
		for _, m := range p.Measurements {
			if m.ID == "" || len(m.Samples) == 0 || (m.Shape.Funding != "none" && m.Shape.Funding != "caller" && m.Shape.Funding != "signature") {
				return errors.New("invalid gas benchmark measurement")
			}
			for _, n := range gasDimensions(m.Shape, true) {
				if n < 0 {
					return errors.New("invalid gas benchmark shape")
				}
			}
			if c := m.Circuit; c != nil && (c.CircuitName == "" || c.Target == "" || !gasDigest.MatchString(c.ArtifactSHA256) || !gasDigest.MatchString(c.VKSHA256)) {
				return errors.New("invalid gas benchmark circuit identity")
			}
			seen := map[string]bool{}
			for _, s := range m.Samples {
				value, valid := new(big.Int).SetString(s.GasUsed, 10)
				hash := strings.ToLower(s.TransactionHash)
				if !valid || !gasDecimal.MatchString(s.GasUsed) || value.Sign() <= 0 || !value.IsUint64() ||
					!gasWord.MatchString(hash) || seen[hash] || !gasWord.MatchString(s.BlockHash) ||
					!gasDecimal.MatchString(s.BlockNumber) || s.CalldataBytes < 0 {
					return errors.New("invalid gas benchmark receipt")
				}
				seen[hash] = true
			}
		}
	}
	return nil
}
func copyGasProfiles(profiles []GasBenchmarkProfile) []GasBenchmarkProfile {
	if profiles == nil {
		return nil
	}
	result := append([]GasBenchmarkProfile{}, profiles...)
	for i := range result {
		if profiles[i].Source.Fingerprints != nil {
			value := *profiles[i].Source.Fingerprints
			result[i].Source.Fingerprints = &value
		}
		if profiles[i].Source.Toolchain != nil {
			value := *profiles[i].Source.Toolchain
			result[i].Source.Toolchain = &value
		}
		if profiles[i].Source.RuntimeContracts != nil {
			result[i].Source.RuntimeContracts = make(map[string]GasBenchmarkRuntimeContract, len(profiles[i].Source.RuntimeContracts))
			for name, contract := range profiles[i].Source.RuntimeContracts {
				result[i].Source.RuntimeContracts[name] = contract
			}
		}
		result[i].Source.Assumptions = append([]string{}, profiles[i].Source.Assumptions...)
		result[i].Measurements = append([]GasBenchmarkMeasurement{}, profiles[i].Measurements...)
		for j := range result[i].Measurements {
			m := &result[i].Measurements[j]
			m.Samples = append([]GasBenchmarkSample{}, m.Samples...)
			if m.Circuit != nil {
				c := *m.Circuit
				m.Circuit = &c
			}
		}
	}
	return result
}

type gasEstimateRequest struct {
	Deployment GasBenchmarkDeployment
	Shape      GasBenchmarkShape
	Circuit    *GasBenchmarkCircuit
}

func sameGasCircuit(a, b *GasBenchmarkCircuit) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.CircuitName == b.CircuitName && a.Target == b.Target &&
		strings.EqualFold(a.ArtifactSHA256, b.ArtifactSHA256) && strings.EqualFold(a.VKSHA256, b.VKSHA256)
}
func sameGasDeployment(a, b GasBenchmarkDeployment) bool {
	return a.ChainID == b.ChainID && strings.EqualFold(a.ProtocolAddress, b.ProtocolAddress) &&
		strings.EqualFold(a.TokenAddress, b.TokenAddress) && strings.EqualFold(a.RuntimeCodeHash, b.RuntimeCodeHash)
}
func measuredGas(m GasBenchmarkMeasurement) (GasReference, error) {
	total := new(big.Int)
	for _, sample := range m.Samples {
		n, _ := new(big.Int).SetString(sample.GasUsed, 10)
		total.Add(total, n)
	}
	count := big.NewInt(int64(len(m.Samples)))
	total.Add(total, new(big.Int).Sub(count, big.NewInt(1)))
	return gasResult(total.Quo(total, count))
}
func gasResult(units *big.Int) (GasReference, error) {
	if !units.IsUint64() || units.Sign() <= 0 {
		return GasReference{}, errors.New("gas reference exceeds uint64")
	}
	return GasReference{GasUnits: units.Uint64()}, nil
}
func gasDimensions(s GasBenchmarkShape, capacity bool) []int {
	values := []int{s.NoteInputs, s.NoteOutputs, s.SourceTrees, s.CreditInputs, s.CreditOutputs, s.CreditRemainders, s.DeliveryBytes}
	if capacity {
		values = append(values, s.Capacity)
	}
	return values
}

// estimateGasShape mirrors the TS upward-only workload envelope. It is intentionally
// coarse for unmeasured combinations and is not a calibrated bound on actual gas.
func estimateGasShape(request gasEstimateRequest, profiles []GasBenchmarkProfile) (GasReference, error) {
	if request.Shape.Target == "noir-recursive" {
		return GasReference{}, errors.New("estimate the complete aggregate, not an individual recursive member")
	}
	var matching, family []GasBenchmarkMeasurement
	var exact GasReference
	for _, p := range profiles {
		for _, m := range p.Measurements {
			if m.Shape.Kind != request.Shape.Kind || m.Shape.Target != request.Shape.Target {
				continue
			}
			family = append(family, m)
			if sameGasCircuit(m.Circuit, request.Circuit) {
				matching = append(matching, m)
				if m.Shape == request.Shape {
					candidate, err := measuredGas(m)
					if err != nil {
						return GasReference{}, err
					}
					if candidate.GasUnits > exact.GasUnits {
						exact = candidate
					}
				}
			}
		}
	}
	if exact.GasUnits > 0 {
		return exact, nil
	}
	candidates := matching
	if len(candidates) == 0 && strings.Contains(request.Shape.Kind, "credit") {
		return GasReference{}, errors.New("no gas reference for this Credit circuit identity")
	}
	if len(candidates) == 0 {
		candidates = family
	}
	if len(candidates) == 0 {
		return GasReference{}, errors.New("no gas reference for this operation kind and target")
	}
	// Policy variants share a public workload. Combine their observed samples
	// before scaling so a cheaper variant cannot lower a modeled reference.
	var baselines []GasBenchmarkMeasurement
	for _, candidate := range candidates {
		found := false
		for i := range baselines {
			if baselines[i].Shape == candidate.Shape && sameGasCircuit(baselines[i].Circuit, candidate.Circuit) {
				baselines[i].Samples = append(baselines[i].Samples, candidate.Samples...)
				found = true
				break
			}
		}
		if !found {
			candidate.Samples = append([]GasBenchmarkSample{}, candidate.Samples...)
			baselines = append(baselines, candidate)
		}
	}
	var minimum *big.Int
	for _, m := range baselines {
		numerator, denominator := big.NewInt(1), big.NewInt(1)
		current, baseline := gasDimensions(request.Shape, len(matching) == 0), gasDimensions(m.Shape, len(matching) == 0)
		for i := range current {
			n := new(big.Int).Add(big.NewInt(int64(current[i])), big.NewInt(1))
			d := new(big.Int).Add(big.NewInt(int64(baseline[i])), big.NewInt(1))
			if new(big.Int).Mul(n, denominator).Cmp(new(big.Int).Mul(numerator, d)) > 0 {
				numerator, denominator = n, d
			}
		}
		units := new(big.Int)
		for _, sample := range m.Samples {
			n, _ := new(big.Int).SetString(sample.GasUsed, 10)
			if n.Cmp(units) > 0 {
				units.Set(n)
			}
		}
		units.Mul(units, numerator)
		units.Add(units, new(big.Int).Sub(denominator, big.NewInt(1)))
		units.Quo(units, denominator)
		if request.Shape.Funding == "signature" && m.Shape.Funding != "signature" {
			units.Mul(units, big.NewInt(2))
		}
		if request.Shape.Permit && !m.Shape.Permit {
			units.Mul(units, big.NewInt(2))
		}
		if minimum == nil || units.Cmp(minimum) < 0 {
			minimum = units
		}
	}
	return gasResult(minimum)
}

// EstimateGas returns only GasUnits, without proving, store access or network calls.
func (o *Operation) EstimateGas() (GasReference, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	request, err := o.gasRequest()
	if err != nil {
		return GasReference{}, err
	}
	profiles := o.session.options.GasBenchmarks
	if err := validateGasProfiles(profiles); err != nil {
		return GasReference{}, err
	}
	var exact GasReference
	for _, p := range profiles {
		if !sameGasDeployment(p.Deployment, request.Deployment) {
			continue
		}
		for _, m := range p.Measurements {
			if m.Shape == request.Shape && sameGasCircuit(m.Circuit, request.Circuit) {
				candidate, err := measuredGas(m)
				if err != nil {
					return GasReference{}, err
				}
				if candidate.GasUnits > exact.GasUnits {
					exact = candidate
				}
			}
		}
	}
	if exact.GasUnits > 0 {
		return exact, nil
	}
	return estimateGasShape(request, defaultGasProfiles)
}

// EstimateGas is available as soon as preparation completes.
func (p *PreparedTransaction) EstimateGas() (GasReference, error) { return p.operation.EstimateGas() }

func (o *Operation) gasRequest() (gasEstimateRequest, error) {
	p := o.plan
	shape := GasBenchmarkShape{Kind: p.Kind, Target: p.Target, Capacity: p.Capacity,
		NoteInputs: len(p.Inputs), NoteOutputs: len(p.Outputs), CreditInputs: len(p.CreditInputs), Funding: p.FundingMode}
	if shape.Target == "" {
		shape.Target = "evm"
	}
	if p.Kind == "transact" {
		shape.Kind = "transfer"
		if p.Params.WithdrawalAmount != nil && p.Params.WithdrawalAmount.Sign() > 0 {
			shape.Kind = "withdraw"
			if !p.NeedsAuthorization {
				shape.Kind = "sweep"
			}
		}
		shape.SourceTrees = len(p.Params.Sources)
	}
	if p.CreateCredit != nil {
		shape.SourceTrees = len(p.CreateCredit.Sources)
	}
	if p.TopUpCredit != nil {
		shape.SourceTrees = len(p.TopUpCredit.Creation.Sources) + 1
	}
	if p.CancelCredit != nil {
		shape.SourceTrees = 1
	}
	if p.CreditOutput != nil {
		shape.CreditOutputs = 1
	}
	if p.DrawCredit != nil {
		shape.Kind = "draw-credit"
		shape.SourceTrees = len(p.DrawCredit.Sources)
		for _, hash := range p.DrawCredit.RemainderCommitments {
			if hash != ([32]byte{}) {
				shape.CreditRemainders++
			}
		}
		shape.CreditOutputs = shape.CreditRemainders
	}
	shape.Permit = p.Funding.Permit.Deadline != nil && p.Funding.Permit.Deadline.Sign() != 0
	for _, d := range o.deliveries() {
		shape.DeliveryBytes += len(d.Ciphertext) + len(d.Clue)
	}
	request := gasEstimateRequest{Shape: shape, Deployment: GasBenchmarkDeployment{
		ChainID: p.Context.ChainID.String(), ProtocolAddress: p.Context.ProtocolAddress.Hex(),
		TokenAddress: p.Context.TokenAddress.Hex(), RuntimeCodeHash: o.session.protocol.Deployment().Protocol.RuntimeCodeHash.Hex(),
	}}
	if p.Kind != "shield" {
		words, err := o.words()
		if err != nil {
			return request, err
		}
		inputs := make([]common.Hash, len(words))
		for i := range words {
			inputs[i] = words[i]
		}
		statement, err := prover.Statement(o.circuitName(), shape.Target, inputs)
		if err != nil {
			return request, fmt.Errorf("gas reference circuit: %w", err)
		}
		request.Circuit = &GasBenchmarkCircuit{CircuitName: statement.CircuitName, Target: statement.Target, ArtifactSHA256: statement.ArtifactSHA256, VKSHA256: statement.VKSHA256}
	}
	return request, nil
}
