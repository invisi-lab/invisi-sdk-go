package broadcaster

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/invisi-lab/invisi-sdk-go/protocol/settlement"
)

func NewCancelToken() (string, string, error) { return settlement.NewCancelToken() }
func addSchedule(params map[string]any, values []*Schedule) error {
	if len(values) > 1 {
		return errors.New("at most one schedule required")
	}
	if len(values) == 1 && values[0] != nil {
		if e := values[0].Validate(); e != nil {
			return e
		}
		params["schedule"] = values[0]
	}
	return nil
}
func sameSchedule(actual *Schedule, values []*Schedule) bool {
	var expected *Schedule
	if len(values) == 1 {
		expected = values[0]
	}
	return actual == nil && expected == nil || actual != nil && expected != nil && *actual == *expected
}
func (c *Client) CancelTransaction(ctx context.Context, id, token string) (Status, error) {
	var out Status
	b, e := hex.DecodeString(tokenPrefix(token))
	if e != nil || len(b) != 32 || token != "0x"+hex.EncodeToString(b) || id == "" || len(id) > 256 {
		return out, errors.New("invalid cancellation handle")
	}
	if e = c.call(ctx, "broadcaster_cancelTransaction", map[string]any{"id": id, "cancelToken": token}, &out); e != nil {
		return out, e
	}
	return out, checkStatus(out, id)
}
func tokenPrefix(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}
