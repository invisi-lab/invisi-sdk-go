package protocol

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

const MaxBatchItems = 1024
const MaxTransportBytes = 16 << 20

func transportBound(v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if len(raw) > MaxTransportBytes {
		return errors.New("wallet transport exceeds 16 MiB")
	}
	return nil
}
func ValidateRequests(requests []Request) error {
	if len(requests) > MaxBatchItems {
		return errors.New("wallet batch exceeds 1024 items")
	}
	if err := transportBound(requests); err != nil {
		return err
	}
	ids := map[string]bool{}
	owner := ""
	for _, q := range requests {
		if err := ValidateRequest(q); err != nil {
			return err
		}
		if ids[q.ID] {
			return errors.New("duplicate wallet request id")
		}
		ids[q.ID] = true
		if owner != "" && q.OwnerAddress != owner {
			return errors.New("mixed wallet batch owners")
		}
		owner = q.OwnerAddress
	}
	return nil
}

// ValidateResponses validates the complete transport and returns canonical request order.
func ValidateResponses(requests []Request, responses []Response) ([]Response, error) {
	if err := ValidateRequests(requests); err != nil {
		return nil, err
	}
	if len(responses) != len(requests) {
		return nil, errors.New("one wallet response required per request")
	}
	if err := transportBound(responses); err != nil {
		return nil, err
	}
	byID := map[string]Response{}
	for _, r := range responses {
		if _, ok := byID[r.ID]; ok {
			return nil, errors.New("duplicate wallet response id")
		}
		byID[r.ID] = r
	}
	out := make([]Response, len(requests))
	for i, q := range requests {
		r, ok := byID[q.ID]
		if !ok {
			return nil, errors.New("missing wallet response id")
		}
		if err := ValidateResponse(q, r); err != nil {
			return nil, err
		}
		out[i] = r
	}
	return out, nil
}
func validateParams(q Request) error {
	switch q.Method {
	case Public:
		var p struct{}
		return Decode(q.Params, &p)
	case Detect:
		var p DetectParams
		if err := Decode(q.Params, &p); err != nil {
			return err
		}
		if p.Clues == nil || len(p.Clues) > 1024 {
			return errors.New("invalid wallet clues")
		}
		for _, s := range p.Clues {
			if _, err := Bytes(s, 68); err != nil {
				return err
			}
		}
	case Authorize:
		var p AuthorizeParams
		if err := Decode(q.Params, &p); err != nil {
			return err
		}
		_, err := p.Core()
		return err
	case DecryptNote, DecryptCredit:
		var raw struct {
			Context  json.RawMessage `json:"context"`
			Envelope string          `json:"envelope"`
		}
		if err := Decode(q.Params, &raw); err != nil {
			return err
		}
		fields, err := objectFields(raw.Context)
		if err != nil {
			return err
		}
		required := []string{"chainId", "protocolAddress", "tokenAddress", "outputCommitment"}
		if q.Method == DecryptCredit {
			required = append(required, "role")
		}
		if len(fields) != len(required) {
			return errors.New("unexpected wallet delivery context field")
		}
		for _, name := range required {
			if _, ok := fields[name]; !ok {
				return errors.New("missing wallet delivery context field")
			}
		}
		var p DecryptParams
		if err := Decode(q.Params, &p); err != nil {
			return err
		}
		if _, err := p.Context.ProtocolContext.Core(); err != nil {
			return err
		}
		token, err := Element(p.Context.TokenAddress)
		if err != nil || primitives.ElementToBigInt(token).BitLen() > 160 {
			return errors.New("invalid token address")
		}
		commitment, err := Bytes(p.Context.OutputCommitment, 32)
		if err != nil {
			return err
		}
		if _, err = Element(common.BytesToHash(commitment).Big().String()); err != nil {
			return err
		}
		envelope, err := Bytes(p.Envelope, -1)
		if err != nil {
			return err
		}
		minimum, version := 160, byte(1)
		if q.Method == DecryptCredit {
			minimum, version = 485, 0x82
			if p.Context.Role != "owner" && p.Context.Role != "spender" {
				return errors.New("invalid credit role")
			}
		} else if p.Context.Role != "" {
			return errors.New("note context cannot contain role")
		}
		if len(envelope) < minimum || len(envelope) > 4096 || envelope[0] != version {
			return errors.New("invalid wallet envelope")
		}
		if _, err = crypto.DecompressPubkey(envelope[1:34]); err != nil {
			return errors.New("invalid envelope ephemeral key")
		}
	}
	return nil
}
