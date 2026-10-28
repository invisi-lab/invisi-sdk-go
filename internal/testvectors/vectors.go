// Package testvectors decodes independent protocol fixtures for SDK regression tests.
package testvectors

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

//go:embed current-protocol.json
var data []byte

type File struct {
	Context  json.RawMessage
	Fixtures map[string]json.RawMessage
}

func Load(t testing.TB) File {
	t.Helper()
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	return f
}
func Decode[T any](t testing.TB, raw json.RawMessage) T {
	t.Helper()
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		t.Fatal(err)
	}
	var result T
	if err := fill(reflect.ValueOf(&result).Elem(), tree); err != nil {
		t.Fatal(err)
	}
	return result
}

var bigIntType = reflect.TypeOf(big.Int{})

func fill(dst reflect.Value, v any) error {
	if v == nil {
		return nil
	}
	if dst.Kind() == reflect.Pointer {
		dst.Set(reflect.New(dst.Type().Elem()))
		if dst.Type().Elem() == bigIntType {
			n, ok := new(big.Int).SetString(fmt.Sprint(v), 10)
			if !ok {
				return fmt.Errorf("bad fixture integer: %v", v)
			}
			dst.Elem().Set(reflect.ValueOf(*n))
			return nil
		}
		return fill(dst.Elem(), v)
	}
	switch dst.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("expected fixture object")
		}
		for i := 0; i < dst.NumField(); i++ {
			name := dst.Type().Field(i).Name
			for k, x := range m {
				if strings.EqualFold(k, name) {
					if err := fill(dst.Field(i), x); err != nil {
						return fmt.Errorf("%s: %w", name, err)
					}
				}
			}
		}
	case reflect.Array:
		if dst.Type().Elem().Kind() != reflect.Uint8 {
			return fmt.Errorf("unsupported fixture array")
		}
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected fixture hex")
		}
		b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
		if err != nil || len(b) != dst.Len() {
			return fmt.Errorf("bad fixed hex fixture %s", s)
		}
		reflect.Copy(dst, reflect.ValueOf(b))
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("expected hex bytes")
			}
			b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
			if err != nil {
				return err
			}
			dst.SetBytes(b)
			return nil
		}
		list, ok := v.([]any)
		if !ok {
			return fmt.Errorf("expected fixture array")
		}
		dst.Set(reflect.MakeSlice(dst.Type(), len(list), len(list)))
		for i, x := range list {
			if err := fill(dst.Index(i), x); err != nil {
				return err
			}
		}
	case reflect.String:
		dst.SetString(v.(string))
	case reflect.Bool:
		dst.SetBool(v.(bool))
	case reflect.Int, reflect.Int64:
		n, ok := new(big.Int).SetString(fmt.Sprint(v), 10)
		if !ok || !n.IsInt64() {
			return fmt.Errorf("bad signed fixture integer")
		}
		dst.SetInt(n.Int64())
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		n, ok := new(big.Int).SetString(fmt.Sprint(v), 10)
		if !ok || n.Sign() < 0 || !n.IsUint64() {
			return fmt.Errorf("bad unsigned fixture integer")
		}
		dst.SetUint(n.Uint64())
	default:
		return fmt.Errorf("unsupported fixture destination %s", dst.Type())
	}
	return nil
}
