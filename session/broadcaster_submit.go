package session

import (
	"context"

	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
)

// SubmitBroadcaster preserves reservations after any potentially externalized request.
func (o *Operation) SubmitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	return o.submitBroadcaster(ctx, c)
}
