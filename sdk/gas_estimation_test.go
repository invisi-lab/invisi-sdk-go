package sdk

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// BB 6 circuit keys do not match the historical BB 5 measurements, so
// preparation uses the existing conservative family reference.
func TestPreparedGasReferenceBeforeProvingWithoutStore(t *testing.T) {
	invisi, input, backend := explicitFixture(t)
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := prepared.EstimateGas()
	if err != nil {
		t.Fatal(err)
	}
	if reference.GasUnits != 1631799 {
		t.Fatalf("TS reference mismatch: %+v", reference)
	}
	if len(backend.requests) != 0 {
		t.Fatal("estimation invoked prover")
	}
	raw, _ := json.Marshal(reference)
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields)
	if len(fields) != 1 || fields["gasUnits"] == nil {
		t.Fatalf("unexpected public fields: %s", raw)
	}
}

func TestManagedGasReferenceMatchesExplicit(t *testing.T) {
	session, token, _ := testSession(t)
	address, _ := session.wallet.GetAddress()
	operation, err := session.PrepareTransfer(context.Background(), TransferPreparation{TokenAddress: token, Outputs: []PaymentOutput{{To: address, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	reference, err := operation.EstimateGas()
	if err != nil || reference.GasUnits != 1631799 {
		t.Fatalf("%+v %v", reference, err)
	}
}

func TestGasReceiptCatalogAndAllDirectKinds(t *testing.T) {
	profiles, err := DefineGasBenchmarks(gasBenchmarkJSON)
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[string]bool{}
	for _, profile := range profiles {
		for _, m := range profile.Measurements {
			kinds[m.Shape.Kind] = true
			if len(m.Samples) != 3 {
				t.Fatal("incomplete receipt evidence")
			}
			request := gasEstimateRequest{Deployment: profile.Deployment, Shape: m.Shape, Circuit: m.Circuit}
			got, err := estimateGasShape(request, profiles)
			if err != nil {
				t.Fatal(err)
			}
			var sum uint64
			for _, sample := range m.Samples {
				value, _ := strconv.ParseUint(sample.GasUsed, 10, 64)
				sum += value
			}
			expected := (sum + 2) / 3
			for _, candidateProfile := range profiles {
				for _, candidate := range candidateProfile.Measurements {
					if candidate.Shape == m.Shape && sameGasCircuit(candidate.Circuit, m.Circuit) {
						mean, _ := measuredGas(candidate)
						if mean.GasUnits > expected {
							expected = mean.GasUnits
						}
					}
				}
			}
			if got.GasUnits != expected {
				t.Fatalf("%s: %+v", m.ID, got)
			}
			request.Deployment.ChainID = "42161"
			next, err := estimateGasShape(request, profiles)
			if err != nil || !reflect.DeepEqual(got, next) {
				t.Fatalf("deployment affected reference: %+v %v", next, err)
			}
			request.Shape.NoteInputs += 2
			larger, err := estimateGasShape(request, profiles)
			if err != nil || larger.GasUnits < got.GasUnits {
				t.Fatalf("invalid larger reference: %+v %v", larger, err)
			}
		}
	}
	if len(kinds) != 8 {
		t.Fatal(kinds)
	}
}

func TestGasReferencesMatchTypeScriptVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/gas-reference-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		TSCommit      string `json:"tsCommit"`
		CatalogSHA256 string `json:"catalogSha256"`
		Cases         []struct {
			ID       string             `json:"id"`
			Request  gasEstimateRequest `json:"request"`
			GasUnits string             `json:"gasUnits"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	if !gasCommit.MatchString(vectors.TSCommit) || vectors.CatalogSHA256 != fmt.Sprintf("%x", sha256.Sum256(gasBenchmarkJSON)) || len(vectors.Cases) < 22 {
		t.Fatal("unexpected TS vector source")
	}
	for _, vector := range vectors.Cases {
		t.Run(vector.ID, func(t *testing.T) {
			got, err := estimateGasShape(vector.Request, defaultGasProfiles)
			if err != nil || strconv.FormatUint(got.GasUnits, 10) != vector.GasUnits {
				t.Fatalf("TS=%s Go=%+v error=%v", vector.GasUnits, got, err)
			}
		})
	}
}

func TestGasRequestMappingForEveryPrepareMethod(t *testing.T) {
	for _, kind := range []string{"shield", "transfer", "withdraw", "sweep", "create-credit", "draw-credit", "top-up-credit", "cancel-credit"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			session, token, backend := testSession(t)
			address, _ := session.wallet.GetAddress()
			var operation *Operation
			var err error
			switch kind {
			case "shield":
				operation, err = session.PrepareShield(ctx, ShieldPreparation{TokenAddress: token, Outputs: []PaymentOutput{{To: address, Amount: big.NewInt(30)}}, Funding: FundingSource{Payer: session.wallet.identity, Amount: big.NewInt(30), Authorization: "caller"}})
			case "transfer":
				operation, err = session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Outputs: []PaymentOutput{{To: address, Amount: big.NewInt(30)}}})
			case "withdraw":
				operation, err = session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Withdrawal: &Withdrawal{To: session.wallet.identity, Amount: big.NewInt(30)}})
			case "sweep":
				operation, err = session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Withdrawal: &Withdrawal{To: session.wallet.identity, Amount: big.NewInt(100)}})
			case "create-credit":
				operation, err = session.PrepareCreateCredit(ctx, CreateCreditPreparation{TokenAddress: token, Spender: address, Amount: big.NewInt(30), Expiry: big.NewInt(1000)})
			default:
				leaf := seedCredit(t, session, token)
				switch kind {
				case "draw-credit":
					operation, err = session.PrepareDrawCredit(ctx, DrawCreditPreparation{TokenAddress: token, Credits: []CreditDrawInput{{Credit: leaf, Amount: big.NewInt(30)}}})
				case "top-up-credit":
					operation, err = session.PrepareTopUpCredit(ctx, TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(30), Expiry: big.NewInt(2000)})
				case "cancel-credit":
					operation, err = session.PrepareCancelCredit(ctx, CancelCreditPreparation{TokenAddress: token, Credit: leaf})
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			request, err := operation.gasRequest()
			if err != nil || request.Shape.Kind != kind {
				t.Fatalf("request=%+v error=%v", request, err)
			}
			reference, err := operation.EstimateGas()
			if err != nil || reference.GasUnits == 0 {
				t.Fatalf("%+v %v", reference, err)
			}
			if len(backend.requests) != 0 {
				t.Fatal("estimation invoked prover")
			}
		})
	}
}

func TestCustomGasProfileIsSnapshottedAndUsedOnlyForExactMatch(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	request, err := prepared.operation.gasRequest()
	if err != nil {
		t.Fatal(err)
	}
	profiles := copyGasProfiles(defaultGasProfiles[:1])
	profiles[0].Deployment = request.Deployment
	measurement := profiles[0].Measurements[3]
	measurement.Shape, measurement.Circuit = request.Shape, request.Circuit
	for i := range measurement.Samples {
		measurement.Samples[i].GasUsed = "123456"
	}
	profiles[0].Measurements = []GasBenchmarkMeasurement{measurement}
	options := invisi.options
	options.GasBenchmarks = profiles
	configured := NewInvisi(options)
	profiles[0].Measurements[0].Samples[0].GasUsed = "1"
	prepared, err = configured.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := prepared.EstimateGas()
	if err != nil || reference.GasUnits != 123456 {
		t.Fatalf("%+v %v", reference, err)
	}
	reference.GasUnits = 1
	again, _ := prepared.EstimateGas()
	if again.GasUnits != 123456 {
		t.Fatal("reference mutation escaped")
	}
	configured.options.GasBenchmarks[0].Deployment.ChainID = "1"
	prepared, err = configured.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	again, err = prepared.EstimateGas()
	if err != nil || again.GasUnits != 1631799 {
		t.Fatalf("unmatched team profile used: %+v %v", again, err)
	}
}

func TestGasCatalogValidationAndUnsupportedTargets(t *testing.T) {
	for _, raw := range []string{"null", "{}", `[{"formatVersion":2}]`} {
		if _, err := DefineGasBenchmarks([]byte(raw)); err == nil {
			t.Fatal("invalid catalog accepted")
		}
	}
	profiles := copyGasProfiles(defaultGasProfiles)
	profiles[0].Measurements[0].Samples[0].GasUsed = "-1"
	raw, _ := json.Marshal(profiles)
	if _, err := DefineGasBenchmarks(raw); err == nil {
		t.Fatal("negative gas accepted")
	}
	m := defaultGasProfiles[0].Measurements[3]
	request := gasEstimateRequest{Shape: m.Shape, Circuit: m.Circuit}
	for _, target := range []string{"noir-recursive", "unknown"} {
		request.Shape.Target = target
		if _, err := estimateGasShape(request, defaultGasProfiles); err == nil {
			t.Fatal("unsupported target accepted")
		}
	}
}

func TestExactGasReferenceUsesHighestMatchingReceiptMean(t *testing.T) {
	profiles := copyGasProfiles(defaultGasProfiles[:1])
	measurement := profiles[0].Measurements[0]
	profiles[0].Measurements = []GasBenchmarkMeasurement{measurement, measurement}
	for i := range profiles[0].Measurements {
		profiles[0].Measurements[i].Samples = append([]GasBenchmarkSample{}, measurement.Samples...)
		for j := range profiles[0].Measurements[i].Samples {
			profiles[0].Measurements[i].Samples[j].GasUsed = strconv.Itoa(100 + i*100)
		}
	}
	request := gasEstimateRequest{Shape: measurement.Shape, Circuit: measurement.Circuit}
	got, err := estimateGasShape(request, profiles)
	if err != nil || got.GasUnits != 200 {
		t.Fatalf("exact reference depends on lower first record: %+v %v", got, err)
	}
	profiles[0].Measurements[0], profiles[0].Measurements[1] = profiles[0].Measurements[1], profiles[0].Measurements[0]
	next, err := estimateGasShape(request, profiles)
	if err != nil || next != got {
		t.Fatalf("exact reference depends on record order: %+v %v", next, err)
	}
}

func TestCreditGasNeverFallsBackToDifferentCircuit(t *testing.T) {
	for _, profile := range defaultGasProfiles {
		for _, measurement := range profile.Measurements {
			if measurement.Circuit == nil || !strings.Contains(measurement.Shape.Kind, "credit") {
				continue
			}
			circuit := *measurement.Circuit
			circuit.ArtifactSHA256 = strings.Repeat("f", 64)
			request := gasEstimateRequest{Shape: measurement.Shape, Circuit: &circuit}
			if _, err := estimateGasShape(request, defaultGasProfiles); err == nil {
				t.Fatalf("%s reused another circuit's Credit receipts", measurement.ID)
			}
		}
	}
}

func TestGasFingerprintMetadataRoundtripAndSnapshot(t *testing.T) {
	profiles := copyGasProfiles(defaultGasProfiles[:1])
	source := &profiles[0].Source
	source.Fingerprints = &GasBenchmarkFingerprints{ProtocolSourceSHA256: strings.Repeat("1", 64), DirtyDiffSHA256: strings.Repeat("2", 64), CompiledArtifactsSHA256: strings.Repeat("3", 64), CreditManifestSHA256: strings.Repeat("4", 64), UserManifestSHA256: strings.Repeat("5", 64), SDKSourceSHA256: strings.Repeat("6", 64)}
	source.Toolchain = &GasBenchmarkToolchain{Solidity: "0.8.38", Barretenberg: "6.0.0", Noir: "1.0.0-rc.6", EVMVersion: "osaka", OptimizerRuns: 200, Anvil: "test"}
	source.BlockGasLimit = "1000000000"
	source.RuntimeContracts = map[string]GasBenchmarkRuntimeContract{"Protocol": {Address: profiles[0].Deployment.ProtocolAddress, RuntimeCodeHash: profiles[0].Deployment.RuntimeCodeHash}}
	raw, err := json.Marshal(profiles)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DefineGasBenchmarks(raw)
	if err != nil || !reflect.DeepEqual(profiles, decoded) {
		t.Fatalf("measurement provenance changed: %v", err)
	}
	snapshot := copyGasProfiles(decoded)
	decoded[0].Source.Fingerprints.DirtyDiffSHA256 = "mutated"
	decoded[0].Source.Toolchain.Solidity = "mutated"
	delete(decoded[0].Source.RuntimeContracts, "Protocol")
	if !reflect.DeepEqual(profiles, snapshot) {
		t.Fatal("measurement provenance mutation escaped snapshot")
	}
}

func TestCurrentCreditGasMatrixAndMemoScaling(t *testing.T) {
	counts := map[string]int{}
	draws := map[int]int{}
	for _, profile := range defaultGasProfiles {
		for _, measurement := range profile.Measurements {
			if !strings.Contains(measurement.Shape.Kind, "credit") {
				continue
			}
			counts[measurement.Shape.Kind]++
			if profile.Source.Fingerprints == nil || profile.Source.Toolchain == nil || profile.Source.Toolchain.Solidity != "0.8.38" || profile.Source.Toolchain.Barretenberg != "6.0.0" {
				t.Fatalf("current receipt provenance missing: %s", measurement.ID)
			}
			request := gasEstimateRequest{Shape: measurement.Shape, Circuit: measurement.Circuit}
			base, err := estimateGasShape(request, defaultGasProfiles)
			if err != nil {
				t.Fatal(err)
			}
			if request.Shape.DeliveryBytes > 0 {
				request.Shape.DeliveryBytes += 256
				memo, err := estimateGasShape(request, defaultGasProfiles)
				if err != nil || memo.GasUnits <= base.GasUnits {
					t.Fatalf("memo did not scale upward: %s base=%d memo=%d err=%v", measurement.ID, base.GasUnits, memo.GasUnits, err)
				}
			}
			if measurement.Shape.Kind == "draw-credit" {
				draws[measurement.Shape.Capacity]++
			}
		}
	}
	if counts["create-credit"] < 1 || counts["draw-credit"] != 8 || counts["top-up-credit"] != 2 || counts["cancel-credit"] != 1 || !reflect.DeepEqual(draws, map[int]int{1: 2, 12: 2, 50: 2, 120: 2}) {
		t.Fatalf("incomplete current Credit receipt matrix: counts=%v draw capacities=%v", counts, draws)
	}
}

func TestModeledGasUsesHighestSampleAcrossSameShapeVariants(t *testing.T) {
	profiles := copyGasProfiles(defaultGasProfiles[:1])
	measurement := profiles[0].Measurements[0]
	profiles[0].Measurements = []GasBenchmarkMeasurement{measurement, measurement}
	for i := range profiles[0].Measurements {
		profiles[0].Measurements[i].Samples = append([]GasBenchmarkSample{}, measurement.Samples...)
		for j := range profiles[0].Measurements[i].Samples {
			profiles[0].Measurements[i].Samples[j].GasUsed = strconv.Itoa(100 + i*100)
		}
	}
	request := gasEstimateRequest{Shape: measurement.Shape, Circuit: measurement.Circuit}
	request.Shape.Funding = "signature"
	got, err := estimateGasShape(request, profiles)
	if err != nil || got.GasUnits != 400 {
		t.Fatalf("modeled reference picked cheaper duplicate baseline: %+v %v", got, err)
	}
	profiles[0].Measurements[0], profiles[0].Measurements[1] = profiles[0].Measurements[1], profiles[0].Measurements[0]
	next, err := estimateGasShape(request, profiles)
	if err != nil || next != got {
		t.Fatalf("modeled reference depends on order: %+v %v", next, err)
	}
}

func TestGasCatalogRejectsMalformedOptionalProvenance(t *testing.T) {
	for _, mutate := range []func(*GasBenchmarkSource){
		func(source *GasBenchmarkSource) { source.Fingerprints = &GasBenchmarkFingerprints{} },
		func(source *GasBenchmarkSource) { source.Toolchain = &GasBenchmarkToolchain{} },
		func(source *GasBenchmarkSource) { source.BlockGasLimit = "-1" },
		func(source *GasBenchmarkSource) {
			source.RuntimeContracts = map[string]GasBenchmarkRuntimeContract{"Protocol": {Address: "invalid", RuntimeCodeHash: "invalid"}}
		},
	} {
		profiles := copyGasProfiles(defaultGasProfiles[:1])
		mutate(&profiles[0].Source)
		raw, _ := json.Marshal(profiles)
		if _, err := DefineGasBenchmarks(raw); err == nil {
			t.Fatal("malformed measurement provenance accepted")
		}
	}
}

func TestHistoricalFundsGasReceiptsRemainUnchanged(t *testing.T) {
	var measurements []GasBenchmarkMeasurement
	for _, profile := range defaultGasProfiles {
		for _, measurement := range profile.Measurements {
			if !strings.Contains(measurement.Shape.Kind, "credit") {
				measurements = append(measurements, measurement)
			}
		}
	}
	raw, err := json.Marshal(measurements)
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 9 || fmt.Sprintf("%x", sha256.Sum256(raw)) != "c252b4d35a79df3b8b56113a4d7635adbfcb401636aa8fa42a8e37d972ba6b04" {
		t.Fatal("historical Funds receipt evidence changed")
	}
}
