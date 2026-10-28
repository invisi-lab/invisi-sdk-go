package receipts

import (
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// eventJSON keeps EVM integers lossless for JavaScript consumers and bytes in hex.
func eventJSON(value any) any {
	switch v := value.(type) {
	case *big.Int:
		return v.String()
	case common.Address:
		return v.Hex()
	case common.Hash:
		return v.Hex()
	case [32]byte:
		return common.Hash(v).Hex()
	case []byte:
		return hexutil.Encode(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[k] = eventJSON(item)
		}
		return out
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		out := make([]any, v.Len())
		for i := range out {
			out[i] = eventJSON(v.Index(i).Interface())
		}
		return out
	case reflect.Struct:
		out := make(map[string]any, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			name := field.Name
			name = strings.ToLower(name[:1]) + name[1:]
			out[name] = eventJSON(v.Field(i).Interface())
		}
		return out
	}
	return value
}
