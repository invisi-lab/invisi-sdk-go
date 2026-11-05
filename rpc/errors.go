package rpc

import (
	"encoding/json"
	"fmt"
)

type ProtocolRpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *ProtocolRpcError) Error() string {
	return fmt.Sprintf("protocol RPC (%d): %s", e.Code, e.Message)
}

type HTTPError struct {
	StatusCode int
	RetryAfter string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("protocol RPC HTTP %d", e.StatusCode) }
