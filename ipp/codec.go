package ipp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// Encode serializes an IPPRequest into a string using the given format.
// JSON format produces "ipp1:{json}".
// Compact format produces "ipp1:{base64url-encoded-msgpack}".
func Encode(request IPPRequest, format Format) (string, error) {
	if err := ValidateIPPRequest(request); err != nil {
		return "", err
	}

	prefix := fmt.Sprintf("ipp%d:", request.V)

	switch format {
	case FormatJSON:
		data, err := json.Marshal(request)
		if err != nil {
			return "", fmt.Errorf("ipp: json marshal: %w", err)
		}
		return prefix + string(data), nil

	case FormatCompact:
		data, err := msgpack.Marshal(request)
		if err != nil {
			return "", fmt.Errorf("ipp: msgpack marshal: %w", err)
		}
		encoded := base64.RawURLEncoding.EncodeToString(data)
		return prefix + encoded, nil

	default:
		return "", fmt.Errorf("ipp: unknown format: %q", format)
	}
}

// Decode parses an encoded IPP string back into an IPPRequest.
// The format is auto-detected from the payload (JSON starts with '{').
func Decode(encoded string) (IPPRequest, error) {
	var zero IPPRequest

	// Parse prefix "ippN:"
	if !strings.HasPrefix(encoded, "ipp") {
		return zero, fmt.Errorf("ipp: invalid prefix: expected 'ipp'")
	}

	colonIdx := strings.Index(encoded, ":")
	if colonIdx < 0 {
		return zero, fmt.Errorf("ipp: missing ':' separator")
	}

	versionStr := encoded[3:colonIdx]
	var version int
	if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
		return zero, fmt.Errorf("ipp: invalid version in prefix: %q", versionStr)
	}

	payload := encoded[colonIdx+1:]

	var request IPPRequest

	if len(payload) > 0 && payload[0] == '{' {
		// JSON format
		if err := json.Unmarshal([]byte(payload), &request); err != nil {
			return zero, fmt.Errorf("ipp: json unmarshal: %w", err)
		}
	} else {
		// Compact format (base64url-encoded msgpack)
		data, err := base64.RawURLEncoding.DecodeString(payload)
		if err != nil {
			return zero, fmt.Errorf("ipp: base64 decode: %w", err)
		}
		if err := msgpack.Unmarshal(data, &request); err != nil {
			return zero, fmt.Errorf("ipp: msgpack unmarshal: %w", err)
		}
	}

	if request.V != version {
		return zero, fmt.Errorf("ipp: version mismatch: prefix=%d, payload=%d", version, request.V)
	}

	if err := ValidateIPPRequest(request); err != nil {
		return zero, err
	}

	return request, nil
}
