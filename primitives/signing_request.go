package primitives

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// SigningRequest is portable JSON EIP-712 data. Integer values are decimal strings.
// Types excludes EIP712Domain so external signing transports can construct it themselves.
type SigningRequest struct {
	Domain      map[string]string         `json:"domain"`
	PrimaryType string                    `json:"primaryType"`
	Types       apitypes.Types            `json:"types"`
	Message     apitypes.TypedDataMessage `json:"message"`
}

// NewSigningRequest validates through the standard EIP-712 encoder and copies public data.
func NewSigningRequest(data apitypes.TypedData) (SigningRequest, error) {
	if _, _, err := apitypes.TypedDataAndHash(data); err != nil {
		return SigningRequest{}, err
	}
	types := apitypes.Types{}
	for name, fields := range data.Types {
		if name != "EIP712Domain" {
			types[name] = append([]apitypes.Type(nil), fields...)
		}
	}
	message, err := normalizeSigningMessage(data.Message, data.PrimaryType, types)
	if err != nil {
		return SigningRequest{}, err
	}
	domain := map[string]string{}
	if data.Domain.Name != "" {
		domain["name"] = data.Domain.Name
	}
	if data.Domain.Version != "" {
		domain["version"] = data.Domain.Version
	}
	if data.Domain.ChainId != nil {
		domain["chainId"] = (*big.Int)(data.Domain.ChainId).String()
	}
	if data.Domain.VerifyingContract != "" {
		domain["verifyingContract"] = data.Domain.VerifyingContract
	}
	if data.Domain.Salt != "" {
		domain["salt"] = data.Domain.Salt
	}
	return SigningRequest{domain, data.PrimaryType, types, message}, nil
}

// TypedData reconstructs validated EIP-712 data after a JSON round trip.
func (r SigningRequest) TypedData() (apitypes.TypedData, error) {
	domain := apitypes.TypedDataDomain{Name: r.Domain["name"], Version: r.Domain["version"], VerifyingContract: r.Domain["verifyingContract"], Salt: r.Domain["salt"]}
	if raw, ok := r.Domain["chainId"]; ok {
		n, ok := new(big.Int).SetString(raw, 10)
		if !ok || n.Sign() < 0 || n.BitLen() > 256 {
			return apitypes.TypedData{}, fmt.Errorf("invalid signing request chainId")
		}
		domain.ChainId = (*ethmath.HexOrDecimal256)(n)
	}
	types := apitypes.Types{}
	for name, fields := range r.Types {
		if name == "EIP712Domain" {
			return apitypes.TypedData{}, fmt.Errorf("signing request types must omit EIP712Domain")
		}
		types[name] = append([]apitypes.Type(nil), fields...)
	}
	fields := []apitypes.Type{}
	for _, f := range []apitypes.Type{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}, {Name: "salt", Type: "bytes32"}} {
		if _, ok := r.Domain[f.Name]; ok {
			fields = append(fields, f)
		}
	}
	for key := range r.Domain {
		if key != "name" && key != "version" && key != "chainId" && key != "verifyingContract" && key != "salt" {
			return apitypes.TypedData{}, fmt.Errorf("unknown signing domain field %s", key)
		}
	}
	types["EIP712Domain"] = fields
	message, err := normalizeSigningMessage(r.Message, r.PrimaryType, types)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	data := apitypes.TypedData{Domain: domain, PrimaryType: r.PrimaryType, Types: types, Message: message}
	if _, _, err = apitypes.TypedDataAndHash(data); err != nil {
		return apitypes.TypedData{}, err
	}
	return data, nil
}
func normalizeSigningMessage(message map[string]interface{}, name string, types apitypes.Types) (apitypes.TypedDataMessage, error) {
	fields, ok := types[name]
	if !ok {
		return nil, fmt.Errorf("unknown signing type %s", name)
	}
	out := apitypes.TypedDataMessage{}
	if len(message) != len(fields) {
		return nil, fmt.Errorf("signing message fields do not match %s", name)
	}
	for _, field := range fields {
		value, ok := message[field.Name]
		if !ok {
			return nil, fmt.Errorf("missing signing field %s", field.Name)
		}
		normalized, err := normalizeSigningValue(value, field.Type, types)
		if err != nil {
			return nil, err
		}
		out[field.Name] = normalized
	}
	return out, nil
}
func normalizeSigningValue(value interface{}, kind string, types apitypes.Types) (interface{}, error) {
	if strings.HasSuffix(kind, "]") {
		start := strings.LastIndex(kind, "[")
		if start < 0 {
			return nil, fmt.Errorf("invalid array type")
		}
		v := reflect.ValueOf(value)
		if !v.IsValid() || (v.Kind() != reflect.Array && v.Kind() != reflect.Slice) {
			return nil, fmt.Errorf("signing array required")
		}
		out := make([]interface{}, v.Len())
		for i := range out {
			item, err := normalizeSigningValue(v.Index(i).Interface(), kind[:start], types)
			if err != nil {
				return nil, err
			}
			out[i] = item
		}
		return out, nil
	}
	if _, ok := types[kind]; ok {
		message, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("signing struct required")
		}
		return normalizeSigningMessage(message, kind, types)
	}
	if strings.HasPrefix(kind, "uint") || strings.HasPrefix(kind, "int") {
		var raw string
		switch n := value.(type) {
		case *big.Int:
			if n == nil {
				return nil, fmt.Errorf("nil signing integer")
			}
			raw = n.String()
		case big.Int:
			raw = n.String()
		case string:
			raw = n
		case json.Number:
			raw = n.String()
		case *ethmath.HexOrDecimal256:
			if n == nil {
				return nil, fmt.Errorf("nil signing integer")
			}
			raw = (*big.Int)(n).String()
		default:
			v := reflect.ValueOf(value)
			if !v.IsValid() {
				return nil, fmt.Errorf("nil signing value")
			}
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				raw = fmt.Sprint(v.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				raw = fmt.Sprint(v.Uint())
			default:
				return nil, fmt.Errorf("unsupported signing integer %T", value)
			}
		}
		n, ok := new(big.Int).SetString(raw, 10)
		if !ok {
			n, ok = new(big.Int).SetString(raw, 0)
		}
		if !ok {
			return nil, fmt.Errorf("invalid signing integer")
		}
		return n.String(), nil
	}
	if b, ok := value.([]byte); ok {
		return "0x" + hex.EncodeToString(b), nil
	}
	switch value.(type) {
	case string, bool:
		return value, nil
	}
	return nil, fmt.Errorf("unsupported signing value %T", value)
}
