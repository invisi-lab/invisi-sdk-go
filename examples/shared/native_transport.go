// Package shared contains public-only example discovery and JSON host transport.
package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

// NativeWalletClient exchanges JSON with the application's external native wallet.
// It contains no keys and performs no cryptographic execution.
type NativeWalletClient struct {
	Encoder *json.Encoder
	Decoder *json.Decoder
}

func (c NativeWalletClient) Execute(ctx context.Context, requests []wallet.Request) ([]wallet.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := wallet.ValidateRequests(requests); err != nil {
		return nil, err
	}
	if requests == nil {
		requests = []wallet.Request{}
	}
	if err := c.Encoder.Encode(requests); err != nil {
		return nil, err
	}
	var raw json.RawMessage
	if err := c.Decoder.Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw) > wallet.MaxTransportBytes || len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '[' {
		return nil, errors.New("native response must be a bounded JSON array")
	}
	var responses []wallet.Response
	if err := json.Unmarshal(raw, &responses); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return wallet.ValidateResponses(requests, responses)
}
