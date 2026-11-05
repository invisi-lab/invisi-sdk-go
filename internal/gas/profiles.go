package gas

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

//go:embed data/gas-benchmarks.json
var GasBenchmarkJSON []byte

var DefaultGasProfiles = func() []GasBenchmarkProfile {
	profiles, err := DefineGasBenchmarks(GasBenchmarkJSON)
	if err != nil {
		panic(err)
	}
	return profiles
}()

var GasWord = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)

var GasDigest = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

var GasCommit = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

var GasDecimal = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

// DefineGasBenchmarks parses trusted JSON receipt records in the same format as TS.
// Validation checks their format, not their publisher's honesty.
func DefineGasBenchmarks(data []byte) ([]GasBenchmarkProfile, error) {
	var profiles []GasBenchmarkProfile
	if err := json.Unmarshal(data, &profiles); err != nil || profiles == nil {
		return nil, errors.New("invalid gas benchmark catalog")
	}
	if err := ValidateGasProfiles(profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

func ValidateGasProfiles(profiles []GasBenchmarkProfile) error {
	for _, p := range profiles {
		if p.FormatVersion != 1 || p.ID == "" || !GasDecimal.MatchString(p.Deployment.ChainID) ||
			!common.IsHexAddress(p.Deployment.ProtocolAddress) || !common.IsHexAddress(p.Deployment.TokenAddress) ||
			!GasWord.MatchString(p.Deployment.RuntimeCodeHash) || p.Source.Verifier != "real" ||
			!GasCommit.MatchString(p.Source.ProtocolCommit) || p.Source.MeasuredAt == "" ||
			p.Source.Assumptions == nil || len(p.Measurements) == 0 {
			return errors.New("invalid gas benchmark profile")
		}
		if f := p.Source.Fingerprints; f != nil {
			for _, fingerprint := range []string{f.ProtocolSourceSHA256, f.DirtyDiffSHA256, f.CompiledArtifactsSHA256, f.CreditManifestSHA256, f.UserManifestSHA256, f.SDKSourceSHA256} {
				if !GasDigest.MatchString(fingerprint) {
					return errors.New("invalid gas benchmark source fingerprint")
				}
			}
		}
		if c := p.Source.Toolchain; c != nil && (c.Solidity == "" || c.Barretenberg == "" || c.Noir == "" || c.EVMVersion == "" || c.OptimizerRuns < 0 || c.Anvil == "") {
			return errors.New("invalid gas benchmark toolchain")
		}
		if p.Source.BlockGasLimit != "" {
			limit, valid := new(big.Int).SetString(p.Source.BlockGasLimit, 10)
			if !valid || !GasDecimal.MatchString(p.Source.BlockGasLimit) || limit.Sign() <= 0 {
				return errors.New("invalid gas benchmark block gas limit")
			}
		}
		for name, contract := range p.Source.RuntimeContracts {
			if name == "" || !common.IsHexAddress(contract.Address) || !GasWord.MatchString(contract.RuntimeCodeHash) {
				return errors.New("invalid gas benchmark runtime contract")
			}
		}
		for _, m := range p.Measurements {
			if m.ID == "" || len(m.Samples) == 0 || (m.Shape.Funding != "none" && m.Shape.Funding != "caller" && m.Shape.Funding != "signature") {
				return errors.New("invalid gas benchmark measurement")
			}
			for _, n := range GasDimensions(m.Shape, true) {
				if n < 0 {
					return errors.New("invalid gas benchmark shape")
				}
			}
			if c := m.Circuit; c != nil && (c.CircuitName == "" || c.Target == "" || !GasDigest.MatchString(c.ArtifactSHA256) || !GasDigest.MatchString(c.VKSHA256)) {
				return errors.New("invalid gas benchmark circuit identity")
			}
			seen := map[string]bool{}
			for _, s := range m.Samples {
				value, valid := new(big.Int).SetString(s.GasUsed, 10)
				hash := strings.ToLower(s.TransactionHash)
				if !valid || !GasDecimal.MatchString(s.GasUsed) || value.Sign() <= 0 || !value.IsUint64() ||
					!GasWord.MatchString(hash) || seen[hash] || !GasWord.MatchString(s.BlockHash) ||
					!GasDecimal.MatchString(s.BlockNumber) || s.CalldataBytes < 0 {
					return errors.New("invalid gas benchmark receipt")
				}
				seen[hash] = true
			}
		}
	}
	return nil
}

func CopyGasProfiles(profiles []GasBenchmarkProfile) []GasBenchmarkProfile {
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
