package broadcaster

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"strconv"
)

// Confirmation is untrusted relay diagnostics; only primary RPC observation authorizes wallet finality.
type Confirmation struct {
	Level  string                  `json:"level"`
	Status protocol.FinalityStatus `json:"status"`
	Block  *ConfirmationBlock      `json:"block"`
}
type ConfirmationBlock struct {
	Number    string      `json:"number"`
	Hash      common.Hash `json:"hash"`
	Timestamp uint64      `json:"timestamp"`
}

func (c *Confirmation) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || len(fields) != 3 || fields["level"] == nil || fields["status"] == nil || fields["block"] == nil {
		return errors.New("invalid confirmation fields")
	}
	if string(fields["block"]) != "null" {
		var block map[string]json.RawMessage
		if json.Unmarshal(fields["block"], &block) != nil || len(block) != 3 || block["number"] == nil || block["hash"] == nil || block["timestamp"] == nil {
			return errors.New("invalid confirmation block fields")
		}
	}
	type plain Confirmation
	var value plain
	if err := strictPublicObject(raw, &value); err != nil {
		return err
	}
	if !value.Status.Valid() {
		return errors.New("invalid broadcaster confirmation status")
	}
	switch value.Level {
	case "pending":
		if value.Block != nil {
			return errors.New("pending confirmation block")
		}
	case "included", "safe", "finalized":
		if value.Block == nil && (value.Status == protocol.FinalityAvailable || value.Level != "included") {
			return errors.New("missing confirmation block")
		}
	default:
		return errors.New("invalid confirmation level")
	}
	if value.Block != nil {
		n, err := strconv.ParseUint(value.Block.Number, 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != value.Block.Number || value.Block.Hash == (common.Hash{}) || value.Block.Timestamp > 1<<53-1 {
			return errors.New("invalid confirmation block")
		}
	}
	*c = Confirmation(value)
	return nil
}

func (s *Status) UnmarshalJSON(raw []byte) error {
	type plain Status
	var value plain
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return errors.New("invalid broadcaster status object")
	}
	if confirmation, present := fields["confirmation"]; present && string(confirmation) == "null" {
		return errors.New("null broadcaster confirmation")
	}
	*s = Status(value)
	return nil
}
