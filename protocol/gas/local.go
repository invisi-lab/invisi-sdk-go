package gas

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

type EstimateResult struct {
	Source         string   `json:"source"`
	Available      bool     `json:"available"`
	GasUnits       *uint64  `json:"gasUnits,omitempty"`
	Reason         string   `json:"reason,omitempty"`
	Statistic      string   `json:"statistic,omitempty"`
	SampleCount    int      `json:"sampleCount,omitempty"`
	ProtocolCommit string   `json:"protocolCommit,omitempty"`
	SourceSHA256   string   `json:"sourceSha256,omitempty"`
	ReferenceID    string   `json:"referenceId,omitempty"`
	Document       string   `json:"document,omitempty"`
	Conditions     string   `json:"conditions,omitempty"`
	Assumptions    []string `json:"assumptions,omitempty"`
	Reference      *Result  `json:"reference,omitempty"`
}

// MarshalJSON 在可用估算中保留零样本数，避免将模型结果误读为收据统计。
func (r EstimateResult) MarshalJSON() ([]byte, error) {
	type encoded EstimateResult
	var sampleCount *int
	if r.Available {
		sampleCount = &r.SampleCount
	}
	return json.Marshal(struct {
		encoded
		SampleCount *int `json:"sampleCount,omitempty"`
	}{encoded(r), sampleCount})
}

//go:embed data/measured-model.json
var measuredJSON []byte
var measured = func() struct {
	ProtocolCommit   string
	SourceSHA256     string
	Capacities       []int
	CreditCapacities []int
	Coefficients     struct {
		Transact []float64
		Draw     []float64
	}
	Rows []struct {
		Shape    Shape
		GasUnits uint64
	}
} {
	var model struct {
		ProtocolCommit   string
		SourceSHA256     string
		Capacities       []int
		CreditCapacities []int
		Coefficients     struct {
			Transact []float64
			Draw     []float64
		}
		Rows []struct {
			Shape    Shape
			GasUnits uint64
		}
	}
	if err := json.Unmarshal(measuredJSON, &model); err != nil {
		panic(err)
	}
	return model
}()

// EstimateGas 保留无 error 的公开入口；需区分无效形状时使用 EstimateGasChecked。
func EstimateGas(shape Shape) EstimateResult {
	result, err := EstimateGasChecked(shape)
	if err != nil {
		return EstimateResult{Source: "benchmark", Reason: "shape-not-covered"}
	}
	return result
}

// EstimateGasChecked 对合法直接执行形状返回本地收据或校准模型，不发起 RPC。
func EstimateGasChecked(query Shape) (EstimateResult, error) {
	unavailable := EstimateResult{Source: "benchmark", Reason: "shape-not-covered"}
	kind, _ := query["kind"].(string)
	switch kind {
	case "shield", "transact", "create-credit", "update-credit", "revoke-credit", "draw-credit":
	default:
		return unavailable, nil
	}
	// JSON 归一化允许调用方传入任意 Go 整数类型，同时拒绝不可序列化的形状。
	raw, err := json.Marshal(query)
	if err != nil {
		return unavailable, err
	}
	var q Shape
	if err = json.Unmarshal(raw, &q); err != nil {
		return unavailable, err
	}
	flag := func(key string, fallback bool) bool {
		v, ok := q[key]
		if !ok {
			return fallback
		}
		return v == true
	}
	count := func(key string, fallback int) (int, error) {
		value, ok := q[key]
		if !ok {
			if fallback >= 0 {
				return fallback, nil
			}
			return 0, fmt.Errorf("invalid gas shape %s", key)
		}
		n, ok := value.(float64)
		if !ok || n < 0 || n > 9007199254740991 || math.Trunc(n) != n {
			return 0, fmt.Errorf("invalid gas shape %s", key)
		}
		return int(n), nil
	}
	capacity := func(tiers []int, size int) (int, error) {
		for _, tier := range tiers {
			if size > 0 && tier >= size {
				if supplied, ok := q["capacity"]; ok && supplied != float64(tier) {
					return 0, fmt.Errorf("unsupported gas shape capacity")
				}
				return tier, nil
			}
		}
		return 0, fmt.Errorf("unsupported gas shape capacity")
	}
	fee := flag("feeEnabled", flag("fee", false))
	shape := Shape{"kind": kind, "fee": fee}
	gas := float64(0)
	inputs, outputs := 0, 0
	switch kind {
	case "shield":
		outputs, err = count("noteOutputs", -1)
		if err != nil {
			return unavailable, err
		}
		if outputs < 1 || outputs > 10 {
			return unavailable, fmt.Errorf("unsupported Shield output count")
		}
		shape["noteOutputs"] = float64(outputs)
	case "transact":
		inputs, err = count("noteInputs", -1)
		if err != nil {
			return unavailable, err
		}
		outputs, err = count("noteOutputs", -1)
		if err != nil {
			return unavailable, err
		}
		tier, e := capacity(measured.Capacities, inputs+outputs)
		if e != nil {
			return unavailable, e
		}
		if inputs == 0 && outputs < 2 {
			return unavailable, fmt.Errorf("single-output public deposits use Shield")
		}
		deposit := flag("deposit", q["scenario"] == "public-funding" || q["scenario"] == "mixed")
		withdrawal := flag("withdrawal", q["scenario"] == "withdrawal" || q["scenario"] == "mixed")
		if inputs == 0 && !deposit || outputs == 0 && !withdrawal && !fee {
			return unavailable, fmt.Errorf("unfunded gas shape")
		}
		fallback := 0
		if inputs > 0 {
			fallback = 1
		}
		trees, e := count("sourceTrees", fallback)
		if e != nil {
			return unavailable, e
		}
		if inputs == 0 && trees != 0 || trees > inputs || inputs > 0 && trees == 0 {
			return unavailable, fmt.Errorf("invalid source tree count")
		}
		shape = Shape{"kind": kind, "fee": fee, "capacity": float64(tier), "noteInputs": float64(inputs), "noteOutputs": float64(outputs), "sourceTrees": float64(trees), "deposit": deposit, "withdrawal": withdrawal}
		features := tierFeatures(measured.Capacities, tier)
		features = append(features, float64(inputs), float64(outputs), boolNumber(outputs > 0), boolNumber(deposit), boolNumber(withdrawal), boolNumber(fee))
		gas = dot(features, measured.Coefficients.Transact) + float64(max(0, trees-1))*10000
	case "draw-credit":
		inputs, err = count("creditInputs", -1)
		if err != nil {
			return unavailable, err
		}
		tier, e := capacity(measured.CreditCapacities, inputs)
		if e != nil {
			return unavailable, e
		}
		payers, e := count("distinctPayers", inputs)
		if e != nil {
			return unavailable, e
		}
		trees, e := count("sourceTrees", 1)
		if e != nil {
			return unavailable, e
		}
		if payers < 1 || payers > inputs || trees < 1 || trees > inputs {
			return unavailable, fmt.Errorf("invalid draw grouping")
		}
		shape = Shape{"kind": kind, "fee": fee, "capacity": float64(tier), "creditInputs": float64(inputs), "distinctPayers": float64(payers), "sourceTrees": float64(trees)}
		features := append(tierFeatures(measured.CreditCapacities, tier), float64(inputs), float64(payers), boolNumber(fee))
		gas = dot(features, measured.Coefficients.Draw) + float64(max(0, trees-1))*10000
	case "revoke-credit":
		authorization := q["authorization"]
		if authorization == nil {
			authorization = "unknown"
		}
		bitmap := q["bitmapWord"]
		if bitmap == nil {
			bitmap = "empty"
		}
		shape["authorization"], shape["bitmapWord"] = authorization, bitmap
	}
	rowIndex := measuredRow(shape)
	if rowIndex >= 0 && kind != "transact" && kind != "draw-credit" {
		gas = float64(measured.Rows[rowIndex].GasUnits)
	} else if rowIndex < 0 && kind == "revoke-credit" {
		auth, bitmap := shape["authorization"], shape["bitmapWord"]
		if (auth != "direct" && auth != "eoa" && auth != "erc1271" && auth != "unknown") || (bitmap != "empty" && bitmap != "nonempty") {
			return unavailable, fmt.Errorf("unsupported credit authorization shape")
		}
		fallback := "eoa"
		if auth == "direct" {
			fallback = "direct"
		}
		idx := measuredRow(Shape{"kind": kind, "fee": fee, "authorization": fallback})
		if idx < 0 {
			return unavailable, fmt.Errorf("missing revoke benchmark")
		}
		gas = float64(measured.Rows[idx].GasUnits)
		if auth == "erc1271" || auth == "unknown" {
			gas += 60000
		}
	}
	noteDeliveries, creditDeliveries := 0, 0
	switch kind {
	case "shield", "transact":
		noteDeliveries = outputs
	case "draw-credit":
		noteDeliveries = 1
		creditDeliveries = 2 * inputs
	case "create-credit", "update-credit":
		creditDeliveries = 2
	}
	extraBytes, e := count("extraDeliveryBytes", 0)
	if e != nil {
		return unavailable, e
	}
	noteBytes, e := count("noteCiphertextBytes", 160)
	if e != nil {
		return unavailable, e
	}
	creditBytes, e := count("creditCiphertextBytes", 485)
	if e != nil {
		return unavailable, e
	}
	clueBytes, e := count("clueBytes", 68)
	if e != nil {
		return unavailable, e
	}
	extra := float64(extraBytes)*32 + float64(noteDeliveries)*float64(max(0, noteBytes-160))*32 + float64(creditDeliveries)*float64(max(0, creditBytes-485))*32 + float64(noteDeliveries+creditDeliveries)*float64(max(0, clueBytes-68))*32
	if flag("permit", false) {
		extra += 70000
	}
	if flag("fundingSignature", false) {
		extra += 40000
	}
	exact := rowIndex >= 0 && extra == 0
	id := kind + "-benchmark-estimate-v1"
	statistic := "estimate"
	samples := 0
	if exact {
		gas = float64(measured.Rows[rowIndex].GasUnits)
		id = "receipt-" + strconv.Itoa(rowIndex)
		statistic = "median"
		samples = 3
	}
	gas = math.Ceil(gas + extra)
	if gas <= 0 || math.IsNaN(gas) || math.IsInf(gas, 0) || gas >= math.Exp2(64) {
		return unavailable, fmt.Errorf("invalid gas model result")
	}
	units := uint64(gas)
	document := "05-benchmarks/5.10-sdk-gas-estimation.md"
	conditions := "Local Anvil31337 / Prague, Solidity0.8.38 optimizer200, real verifiers, reference ERC20."
	assumptions := []string{"Pre-signing local fee preview; not an RPC simulation or transaction gas limit.", "Token behavior, storage warmth, tree cursor and rollover can change actual gas.", "Additional source trees, authorization and delivery bytes use explicit benchmark-calibrated allowances."}
	reference := &Result{Measurement: Measurement{ID: id, Shape: shape, GasUnits: strconv.FormatUint(units, 10), Statistic: statistic, SampleCount: samples}, Source: Source{Path: document, SHA256: measured.SourceSHA256, Conditions: conditions}, Assumptions: append([]string(nil), assumptions...)}
	return EstimateResult{Source: "benchmark", Available: true, GasUnits: &units, Statistic: statistic, SampleCount: samples, ProtocolCommit: measured.ProtocolCommit, SourceSHA256: measured.SourceSHA256, ReferenceID: id, Document: document, Conditions: conditions, Assumptions: assumptions, Reference: reference}, nil
}

func boolNumber(v bool) float64 {
	if v {
		return 1
	}
	return 0
}
func tierFeatures(tiers []int, tier int) []float64 {
	features := make([]float64, len(tiers))
	for i, c := range tiers {
		features[i] = boolNumber(c == tier)
	}
	return features
}
func dot(features, coefficients []float64) float64 {
	sum := float64(0)
	for i, x := range features {
		sum += x * coefficients[i]
	}
	return sum
}
func measuredRow(shape Shape) int {
	for i, row := range measured.Rows {
		match := true
		for key, v := range shape {
			if row.Shape[key] != v {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
