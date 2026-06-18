package ipp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// encodeEnvelope serializes any envelope value using the given format and
// version prefix. Shared by Encode (request) and EncodeVoucher (voucher).
func encodeEnvelope(version int, v any, format Format) (string, error) {
	prefix := fmt.Sprintf("ipp%d:", version)

	switch format {
	case FormatJSON:
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("ipp: json marshal: %w", err)
		}
		return prefix + string(data), nil

	case FormatCompact:
		data, err := msgpack.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("ipp: msgpack marshal: %w", err)
		}
		encoded := base64.RawURLEncoding.EncodeToString(data)
		return prefix + encoded, nil

	default:
		return "", fmt.Errorf("ipp: unknown format: %q", format)
	}
}

// decodeEnvelope parses the "ippN:" prefix and decodes the payload into dst.
// dst must be a pointer. The format is auto-detected ('{' → JSON, else base64url+msgpack).
// Version consistency is checked after decoding via the V field in dst.
func decodeEnvelope(encoded string, dst any) error {
	if !strings.HasPrefix(encoded, "ipp") {
		return fmt.Errorf("ipp: invalid prefix: expected 'ipp'")
	}

	colonIdx := strings.Index(encoded, ":")
	if colonIdx < 0 {
		return fmt.Errorf("ipp: missing ':' separator")
	}

	versionStr := encoded[3:colonIdx]
	var version int
	if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
		return fmt.Errorf("ipp: invalid version in prefix: %q", versionStr)
	}

	payload := encoded[colonIdx+1:]

	if len(payload) > 0 && payload[0] == '{' {
		// JSON format
		if err := json.Unmarshal([]byte(payload), dst); err != nil {
			return fmt.Errorf("ipp: json unmarshal: %w", err)
		}
	} else {
		// Compact format (base64url-encoded msgpack)
		data, err := base64.RawURLEncoding.DecodeString(payload)
		if err != nil {
			return fmt.Errorf("ipp: base64 decode: %w", err)
		}
		if err := msgpack.Unmarshal(data, dst); err != nil {
			return fmt.Errorf("ipp: msgpack unmarshal: %w", err)
		}
	}

	// Verify version consistency using a type-switch on common envelope shapes.
	switch d := dst.(type) {
	case *IPPRequest:
		if d.V != version {
			return fmt.Errorf("ipp: version mismatch: prefix=%d, payload=%d", version, d.V)
		}
	case *IPPVoucher:
		if d.V != version {
			return fmt.Errorf("ipp: version mismatch: prefix=%d, payload=%d", version, d.V)
		}
	default:
		// For header-only peek structs we skip the version check.
	}

	return nil
}

// Encode serializes an IPPRequest into a string using the given format.
// JSON format produces "ipp1:{json}".
// Compact format produces "ipp1:{base64url-encoded-msgpack}".
func Encode(request IPPRequest, format Format) (string, error) {
	if err := ValidateIPPRequest(request); err != nil {
		return "", err
	}
	return encodeEnvelope(request.V, request, format)
}

// Decode parses an encoded IPP string back into an IPPRequest.
// The format is auto-detected from the payload (JSON starts with '{').
// Returns an error if the decoded Type is not MessageTypeRequest.
func Decode(encoded string) (IPPRequest, error) {
	var request IPPRequest
	if err := decodeEnvelope(encoded, &request); err != nil {
		return IPPRequest{}, err
	}

	if request.Type != MessageTypeRequest {
		return IPPRequest{}, fmt.Errorf("ipp: expected type %q, got %q", MessageTypeRequest, request.Type)
	}

	if err := ValidateIPPRequest(request); err != nil {
		return IPPRequest{}, err
	}

	return request, nil
}

// EncodeVoucher mirrors Encode for IPPVoucher (shared ipp{V}: prefix, type inside).
func EncodeVoucher(v IPPVoucher, format Format) (string, error) {
	return encodeEnvelope(v.V, v, format)
}

// DecodeVoucher decodes and asserts Type == "voucher".
func DecodeVoucher(encoded string) (IPPVoucher, error) {
	var v IPPVoucher
	if err := decodeEnvelope(encoded, &v); err != nil {
		return IPPVoucher{}, err
	}
	if v.Type != MessageTypeVoucher {
		return IPPVoucher{}, fmt.Errorf("ipp: expected type %q, got %q", MessageTypeVoucher, v.Type)
	}
	return v, nil
}

// PeekType decodes only the envelope header to read the discriminator.
func PeekType(encoded string) (IPPMessageType, error) {
	var hdr struct {
		Type IPPMessageType `json:"type" msgpack:"type"`
	}
	if err := decodeEnvelope(encoded, &hdr); err != nil {
		return "", err
	}
	return hdr.Type, nil
}
