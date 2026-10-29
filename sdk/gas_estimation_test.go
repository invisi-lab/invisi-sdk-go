package sdk

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"strconv"
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
	for _, m := range profiles[0].Measurements {
		kinds[m.Shape.Kind] = true
		if len(m.Samples) != 3 {
			t.Fatal("incomplete receipt evidence")
		}
		request := gasEstimateRequest{Deployment: profiles[0].Deployment, Shape: m.Shape, Circuit: m.Circuit}
		got, err := estimateGasShape(request, profiles)
		if err != nil {
			t.Fatal(err)
		}
		var sum uint64
		for _, sample := range m.Samples {
			value, _ := strconv.ParseUint(sample.GasUsed, 10, 64)
			sum += value
		}
		if got.GasUnits != (sum+2)/3 {
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
		TSCommit string `json:"tsCommit"`
		Cases    []struct {
			ID       string             `json:"id"`
			Request  gasEstimateRequest `json:"request"`
			GasUnits string             `json:"gasUnits"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	if vectors.TSCommit != "f07f53f" || len(vectors.Cases) != 22 {
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
				operation, err = session.PrepareWithdraw(ctx, WithdrawalPreparation{TokenAddress: token, To: session.wallet.identity, Amount: big.NewInt(30)})
			case "sweep":
				operation, err = session.PrepareSweep(ctx, SweepPreparation{TokenAddress: token})
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
