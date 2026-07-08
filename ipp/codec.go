package ipp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// encodeEnvelope serializes an envelope value. JSON is a bare JSON document so
// it stays valid for any JSON parser / application/json body; compact is
// base64url(msgpack) behind an "ippN:" prefix carrying the protocol version.
// Shared by Encode (request) and EncodeVoucher (voucher).
func encodeEnvelope(version int, v any, format Format) (string, error) {
	switch format {
	case FormatJSON:
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("ipp: json marshal: %w", err)
		}
		return string(data), nil

	case FormatCompact:
		data, err := msgpack.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("ipp: msgpack marshal: %w", err)
		}
		encoded := base64.RawURLEncoding.EncodeToString(data)
		return fmt.Sprintf("ipp%d:%s", version, encoded), nil

	default:
		return "", fmt.Errorf("ipp: unknown format: %q", format)
	}
}

// decodeEnvelope decodes an envelope string into dst (must be a pointer),
// auto-detecting the format from the first byte: '{' is a bare JSON document,
// "ipp" is an "ippN:" compact envelope (base64url(msgpack)). The prefix-vs-payload
// version check applies only to compact; a JSON document carries its version
// solely in the "v" field, so there is no second source to cross-check.
func decodeEnvelope(encoded string, dst any) error {
	switch {
	case strings.HasPrefix(encoded, "{"):
		if err := json.Unmarshal([]byte(encoded), dst); err != nil {
			return fmt.Errorf("ipp: json unmarshal: %w", err)
		}
		return nil

	case strings.HasPrefix(encoded, "ipp"):
		colonIdx := strings.Index(encoded, ":")
		if colonIdx < 0 {
			return fmt.Errorf("ipp: invalid compact format: missing ':' separator")
		}
		versionStr := encoded[3:colonIdx]
		var version int
		if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
			return fmt.Errorf("ipp: invalid version in prefix: %q", versionStr)
		}
		data, err := base64.RawURLEncoding.DecodeString(encoded[colonIdx+1:])
		if err != nil {
			return fmt.Errorf("ipp: base64 decode: %w", err)
		}
		if err := msgpack.Unmarshal(data, dst); err != nil {
			return fmt.Errorf("ipp: msgpack unmarshal: %w", err)
		}
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
			// header-only peek struct: no V field to cross-check
		}
		return nil

	default:
		return fmt.Errorf("ipp: unknown format: expected JSON ('{') or compact ('ipp')")
	}
}

// Encode serializes an IPPRequest into a string using the given format.
// JSON produces a bare JSON document; compact produces "ippN:<base64url(msgpack)>".
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

// EncodeVoucher mirrors Encode for IPPVoucher; request vs voucher is the "type"
// field inside the envelope, not the format or any wire prefix.
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
