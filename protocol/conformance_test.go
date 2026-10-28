package protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// The fixtures are shared byte-for-byte with the current TS SDK. Dummy proof bytes
// exercise encoding and public statements only, never proof soundness.
func decodeReference(t *testing.T, input any, target any) {
	t.Helper()
	if e := referenceValue(input, reflect.ValueOf(target).Elem()); e != nil {
		t.Fatal(e)
	}
}
func referenceValue(input any, out reflect.Value) error {
	if out.Type() == reflect.TypeFor[*big.Int]() {
		text := fmt.Sprint(input)
		n, ok := new(big.Int).SetString(text, 10)
		if !ok {
			return fmt.Errorf("invalid integer %s", text)
		}
		out.Set(reflect.ValueOf(n))
		return nil
	}
	if input == nil {
		return nil
	}
	switch out.Kind() {
	case reflect.Struct:
		values, ok := input.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object")
		}
		for k, v := range values {
			for i := 0; i < out.NumField(); i++ {
				if strings.EqualFold(out.Type().Field(i).Name, k) {
					if e := referenceValue(v, out.Field(i)); e != nil {
						return e
					}
				}
			}
		}
		return nil
	case reflect.Array:
		if text, ok := input.(string); ok {
			b, e := hex.DecodeString(strings.TrimPrefix(text, "0x"))
			if e != nil || len(b) != out.Len() {
				return fmt.Errorf("invalid byte array")
			}
			reflect.Copy(out, reflect.ValueOf(b))
			return nil
		}
	case reflect.Slice:
		if out.Type().Elem().Kind() == reflect.Uint8 {
			b, e := hex.DecodeString(strings.TrimPrefix(input.(string), "0x"))
			if e != nil {
				return e
			}
			out.SetBytes(b)
			return nil
		}
		list := input.([]any)
		out.Set(reflect.MakeSlice(out.Type(), len(list), len(list)))
		for i, v := range list {
			if e := referenceValue(v, out.Index(i)); e != nil {
				return e
			}
		}
		return nil
	case reflect.Uint8, reflect.Uint64:
		value, err := strconv.ParseUint(fmt.Sprint(input), 10, out.Type().Bits())
		if err != nil {
			return err
		}
		out.SetUint(value)
		return nil
	case reflect.String:
		out.SetString(input.(string))
		return nil
	}
	return fmt.Errorf("unsupported fixture target %s", out.Type())
}
func TestCurrentTSProtocolStatementsAndCalldata(t *testing.T) {
	raw, e := os.ReadFile("protocol-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var all struct {
		Context  map[string]any
		Fixtures map[string]map[string]any
	}
	if e = json.Unmarshal(raw, &all); e != nil {
		t.Fatal(e)
	}
	var context TransactionContext
	decodeReference(t, all.Context, &context)
	for _, name := range []string{"transact1", "transact15", "transact57", "transact113", "create", "draw1", "draw2", "draw13", "draw51", "cancel", "topUp", "shield", "single", "crossTwo"} {
		t.Run(name, func(t *testing.T) {
			f := all.Fixtures[name]
			var deliveries []Delivery
			if d := f["deliveries"]; d != nil {
				decodeReference(t, d, &deliveries)
			}
			funding := FundingParams{}
			if d := f["funding"]; d != nil {
				decodeReference(t, d, &funding)
			}
			var inputs [][32]byte
			var call *Calldata
			var err error
			switch {
			case strings.HasPrefix(name, "transact") || name == "single" || name == "crossTwo":
				var p CrossTreeTransactParams
				decodeReference(t, f["params"], &p)
				inputs, err = TransactUserPublicInputs(context, p, deliveries, funding)
				if expected, ok := f["fundingRequestHash"].(string); ok {
					hash, e := TransactFundingRequestHash([]byte{0x12, 0x34}, p, deliveries)
					if e != nil || hash.Hex() != expected {
						t.Fatal("transact funding hash differs", e)
					}
				}
				if name == "transact1" || name == "single" || name == "crossTwo" {
					call, err = BuildTransactCalldata(BuildTransactCalldataParams{Context: context, Proof: []byte{0x12, 0x34}, Params: p, Deliveries: deliveries, Funding: funding})
				}
				if f["userPublicInputs"] != nil {
					f = cloneFixture(f)
					f["publicInputs"] = f["userPublicInputs"]
				} else {
					inputs = nil
				}
			case name == "create":
				var p CreateCreditParams
				decodeReference(t, f["params"], &p)
				inputs, err = CreateCreditPublicInputs(context, p, deliveries, funding)
				hash, e := CreateCreditRequestHash(p, deliveries, funding)
				if e != nil || hash.Hex() != f["requestHash"].(string) {
					t.Fatal("Credit creation request hash differs", e)
				}
				funding.Signature = []byte{0xab, 0xcd}
				call, err = BuildCreateCreditCalldata(BuildCreateCreditCalldataParams{Context: context, Proof: []byte{0x12, 0x34}, Params: p, Deliveries: deliveries, Funding: funding})
			case strings.HasPrefix(name, "draw"):
				var p CreditDrawParams
				decodeReference(t, f["params"], &p)
				inputs, err = CreditDrawPublicInputs(context, p, deliveries)
				if name == "draw1" {
					call, err = BuildCreditDrawCalldata(BuildCreditDrawCalldataParams{Context: context, Proof: []byte{0x12, 0x34}, Params: p, Deliveries: deliveries})
				}
			case name == "cancel":
				var p CancelCreditParams
				decodeReference(t, f["params"], &p)
				inputs, err = CancelCreditPublicInputs(context, p)
				hash, e := CancelCreditRequestHash(p)
				if e != nil || hash.Hex() != f["requestHash"].(string) {
					t.Fatal("Credit cancellation request hash differs", e)
				}
				call, err = BuildCancelCreditCalldata(BuildCancelCreditCalldataParams{Context: context, Proof: []byte{0x12, 0x34}, Params: p})
			case name == "topUp":
				var p TopUpCreditParams
				decodeReference(t, f["params"], &p)
				inputs, err = TopUpCreditPublicInputs(context, p, deliveries, funding)
				hash, e := TopUpCreditRequestHash(p, deliveries, funding)
				if e != nil || hash.Hex() != f["requestHash"].(string) {
					t.Fatal("Credit top-up request hash differs", e)
				}
				funding.Signature = []byte{0xab, 0xcd}
				call, err = BuildTopUpCreditCalldata(BuildTopUpCreditCalldataParams{Context: context, Proof: []byte{0x12, 0x34}, Params: p, Deliveries: deliveries, Funding: funding})
			case name == "shield":
				var p ShieldParams
				var outputs []ShieldOutput
				decodeReference(t, f["params"], &p)
				decodeReference(t, f["outputs"], &outputs)
				hash, e := ShieldFundingRequestHash(p, outputs, deliveries)
				if e != nil || hash.Hex() != f["requestHash"].(string) {
					t.Fatal("shield funding hash differs", e)
				}
				call, err = BuildShieldCalldata(BuildShieldCalldataParams{Context: context, Params: p, Outputs: outputs, Deliveries: deliveries, Funding: funding})
			}
			if err != nil {
				t.Fatal(err)
			}
			if inputs != nil {
				expected := f["publicInputs"].([]any)
				if len(expected) != len(inputs) {
					t.Fatalf("statement width %d != %d", len(inputs), len(expected))
				}
				for i, input := range inputs {
					if common.Hash(input).Hex() != expected[i].(string) {
						t.Fatalf("public input %d differs", i)
					}
				}
			}
			if call != nil && f["calldata"] != nil {
				if "0x"+hex.EncodeToString(call.Data) != f["calldata"].(string) {
					t.Fatal("calldata differs from TS SDK")
				}
			}
		})
	}
}
func cloneFixture(v map[string]any) map[string]any {
	r := map[string]any{}
	for k, v := range v {
		r[k] = v
	}
	return r
}
