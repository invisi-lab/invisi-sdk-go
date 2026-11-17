package ipp

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MaxResponseBytes = 1 << 20

var canonicalNumber = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

// DecodeJSON rejects duplicate keys, malformed Unicode, invalid numeric tokens and trailing data.
func DecodeJSON(data []byte) (any, error) {
	if len(data) > MaxResponseBytes || !utf8.Valid(data) || !validEscapes(data) {
		return nil, invalid()
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	v, e := decodeValue(d, 0)
	if e != nil {
		return nil, invalid()
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, invalid()
	}
	return v, nil
}
func validEscapes(data []byte) bool {
	// encoding/json silently replaces unpaired surrogates. Inspect each raw JSON string first.
	for i := 0; i < len(data); i++ {
		if data[i] != '"' {
			continue
		}
		i++
		for i < len(data) && data[i] != '"' {
			if data[i] != '\\' {
				i++
				continue
			}
			i++
			if i >= len(data) {
				return false
			}
			if data[i] == 'u' {
				if i+4 >= len(data) {
					return false
				}
				n, e := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
				if e != nil {
					return false
				}
				i += 4
				if n >= 0xd800 && n <= 0xdbff {
					if i+6 >= len(data) || string(data[i+1:i+3]) != "\\u" {
						return false
					}
					m, e := strconv.ParseUint(string(data[i+3:i+7]), 16, 16)
					if e != nil || m < 0xdc00 || m > 0xdfff {
						return false
					}
					i += 6
				} else if n >= 0xdc00 && n <= 0xdfff {
					return false
				}
			}
			i++
		}
	}
	return true
}
func decodeValue(d *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, invalid()
	}
	tok, e := d.Token()
	if e != nil {
		return nil, e
	}
	switch v := tok.(type) {
	case json.Delim:
		if v == '{' {
			m := map[string]any{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return nil, e
				}
				s, ok := k.(string)
				if !ok {
					return nil, invalid()
				}
				if _, ok = m[s]; ok {
					return nil, invalid()
				}
				x, e := decodeValue(d, depth+1)
				if e != nil {
					return nil, e
				}
				m[s] = x
			}
			end, e := d.Token()
			if e != nil || end != json.Delim('}') {
				return nil, invalid()
			}
			return m, nil
		}
		if v == '[' {
			a := []any{}
			for d.More() {
				x, e := decodeValue(d, depth+1)
				if e != nil {
					return nil, e
				}
				a = append(a, x)
			}
			end, e := d.Token()
			if e != nil || end != json.Delim(']') {
				return nil, invalid()
			}
			return a, nil
		}
		return nil, invalid()
	case json.Number:
		f, e := strconv.ParseFloat(string(v), 64)
		if e != nil || math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, invalid()
		}
		if f == math.Trunc(f) && f >= 0 && f <= 9007199254740991 {
			if f == 0 {
				return json.Number("0"), nil
			}
			return json.Number(strconv.FormatFloat(f, 'f', -1, 64)), nil
		}
	}
	return tok, nil
}
func object(v any, required string, optional ...string) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, invalid()
	}
	allowed := map[string]bool{}
	for _, k := range strings.Fields(required) {
		allowed[k] = true
		if _, ok = m[k]; !ok {
			return nil, invalid()
		}
	}
	for _, s := range optional {
		for _, k := range strings.Fields(s) {
			allowed[k] = true
		}
	}
	for k := range m {
		if !allowed[k] {
			return nil, invalid()
		}
	}
	return m, nil
}
func decodeTyped[T any](m any) (*T, error) {
	b, e := json.Marshal(m)
	if e != nil {
		return nil, invalid()
	}
	var t T
	if json.Unmarshal(b, &t) != nil {
		return nil, invalid()
	}
	return &t, nil
}

// EncodeJSON rejects invalid Go UTF-8 strings before encoding/json can replace them.
// It also enforces the same bounded, unambiguous JSON representation as DecodeJSON.
func EncodeJSON(value any) ([]byte, error) {
	if !validGoStrings(reflect.ValueOf(value), 0) {
		return nil, invalid()
	}
	data, e := json.Marshal(value)
	if e != nil {
		return nil, invalid()
	}
	if _, e = DecodeJSON(data); e != nil {
		return nil, e
	}
	return data, nil
}
func validGoStrings(v reflect.Value, depth int) bool {
	if depth > 64 {
		return false
	}
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String:
		return utf8.ValidString(v.String())
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return validGoStrings(v.Elem(), depth+1)
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if !validGoStrings(v.Index(i), depth+1) {
				return false
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !validGoStrings(v.Field(i), depth+1) {
				return false
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if !validGoStrings(iter.Key(), depth+1) || !validGoStrings(iter.Value(), depth+1) {
				return false
			}
		}
	}
	return true
}
