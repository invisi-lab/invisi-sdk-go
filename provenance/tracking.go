package provenance

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrTrackingStaleGeneration = errors.New("TRACKING_STALE_GENERATION")

type TrackingMatch struct {
	TxHash         string `json:"txHash"`
	BlockHash      string `json:"blockHash"`
	OperationIndex string `json:"operationIndex"`
}
type Tracking struct {
	ID                 string         `json:"id"`
	Status             string         `json:"status"`
	ExpiresAt          string         `json:"expiresAt"`
	OperationDigest    string         `json:"operationDigest"`
	Matched            *TrackingMatch `json:"matched,omitempty"`
	ReasonCode         string         `json:"reasonCode,omitempty"`
	CheckpointSequence string         `json:"checkpointSequence,omitempty"`
}

func TrackingFromResult(result Result, id string) (Tracking, error) {
	var t Tracking
	d := json.NewDecoder(strings.NewReader(string(result["tracking"])))
	d.DisallowUnknownFields()
	if e := d.Decode(&t); e != nil {
		return t, e
	}
	if !regexp.MustCompile(`^[a-f0-9]{48}$`).MatchString(t.ID) || (id != "" && t.ID != id) {
		return t, errors.New("invalid tracking id")
	}
	digest, e := word(t.OperationDigest)
	if e != nil || digest != t.OperationDigest {
		return t, errors.New("invalid tracking digest")
	}
	expires, e := Uint(t.ExpiresAt, 64)
	if e != nil || expires.Sign() == 0 || expires.String() != t.ExpiresAt {
		return t, errors.New("invalid tracking expiry")
	}
	switch t.Status {
	case "stale_generation":
		return t, ErrTrackingStaleGeneration
	case "pending_confirmation":
	case "pending_source", "recognized":
		if t.Matched == nil {
			return t, errors.New("tracking locator required")
		}
	case "expired", "superseded":
	default:
		return t, errors.New("unknown tracking status")
	}
	if t.Matched != nil {
		m := t.Matched
		for _, v := range []string{m.TxHash, m.BlockHash} {
			w, e := word(v)
			if e != nil || w != v {
				return t, errors.New("invalid tracking locator")
			}
		}
		n, e := Uint(m.OperationIndex, 32)
		if e != nil || n.String() != m.OperationIndex {
			return t, errors.New("invalid tracking index")
		}
	}
	return t, nil
}
func (c *Client) GetTracking(ctx context.Context, id string) (Result, error) {
	if !regexp.MustCompile(`^[a-f0-9]{48}$`).MatchString(id) {
		return nil, errors.New("invalid tracking id")
	}
	r, e := c.call(ctx, "provenance_getTracking", c.publicQuery(map[string]any{"trackingId": id}))
	if e != nil {
		return nil, e
	}
	_, e = TrackingFromResult(r, id)
	return r, e
}
func (c *Client) WaitTracking(ctx context.Context, id string, interval, timeout time.Duration) (Result, error) {
	if interval <= 0 || interval > time.Minute || timeout <= 0 {
		return nil, errors.New("tracking polling limits")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		r, e := c.GetTracking(ctx, id)
		if e != nil {
			return nil, e
		}
		tracking, e := TrackingFromResult(r, id)
		if e != nil {
			return nil, e
		}
		if tracking.Status == "recognized" || tracking.Status == "expired" || tracking.Status == "superseded" {
			return r, nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
func PreflightTracking(result Result, index, digest string) (Tracking, error) {
	var decision string
	json.Unmarshal(result["broadcastDecision"], &decision)
	if decision != "allow" {
		return Tracking{}, errors.New("preflight candidate was not allowed")
	}
	member := result
	if data, ok := result["operations"]; ok {
		var members []Result
		if e := json.Unmarshal(data, &members); e != nil {
			return Tracking{}, e
		}
		member = nil
		for _, candidate := range members {
			var i string
			json.Unmarshal(candidate["operationIndex"], &i)
			if i == index {
				member = candidate
				break
			}
		}
	}
	if member == nil {
		return Tracking{}, errors.New("preflight member missing")
	}
	var actual string
	json.Unmarshal(member["operationDigest"], &actual)
	json.Unmarshal(member["broadcastDecision"], &decision)
	if decision != "allow" || !equalNumber(actual, digest) {
		return Tracking{}, errors.New("preflight statement mismatch")
	}
	t, e := TrackingFromResult(member, "")
	if e != nil {
		return t, e
	}
	if t.Status != "pending_confirmation" && t.Status != "pending_source" && t.Status != "recognized" {
		return t, errors.New("tracking no longer permits broadcast")
	}
	if !equalNumber(t.OperationDigest, digest) {
		return t, errors.New("tracking statement mismatch")
	}
	if _, e := strconv.ParseUint(index, 10, 32); e != nil {
		return t, e
	}
	return t, nil
}
