package rpc

import (
	"context"
	"errors"
)

// GetHistoryUpdates reads both discovery streams with one independently checked
// deployment/status/finality observation. The observation is local to this call;
// both streams retain their numeric canonical anchor checks. No consumption,
// revocation, membership or future operation reads are cached by this method.
func (c *ProtocolRpcClient) GetHistoryUpdates(ctx context.Context, scope DeploymentContext, r ProtocolUpdateRange) (ProtocolNoteUpdates, ProtocolCreditUpdates, error) {
	status, err := c.readStatus(ctx, scope, true)
	if err != nil {
		return ProtocolNoteUpdates{}, ProtocolCreditUpdates{}, err
	}
	notes, err := c.getNoteUpdatesWithStatus(ctx, scope, r, &status)
	if err != nil {
		return ProtocolNoteUpdates{}, ProtocolCreditUpdates{}, err
	}
	credits, err := c.getCreditTransitionsWithStatus(ctx, scope, r, &status)
	if err != nil {
		return ProtocolNoteUpdates{}, ProtocolCreditUpdates{}, err
	}
	if notes.Block != credits.Block {
		return ProtocolNoteUpdates{}, ProtocolCreditUpdates{}, errors.New("history discovery anchors differ")
	}
	return notes, credits, nil
}
