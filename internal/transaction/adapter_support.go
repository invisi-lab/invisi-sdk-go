// Package-internal inspection helpers support invariant tests across the public package boundary.
// They are deliberately not re-exported by sdk or session.
package transaction

import (
	"context"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func PreparedExecution(p *PreparedTransaction) *Execution { return p.operation }
func OutputSnapshot(n NoteOutput) *OutputIntent           { return n.saved }
func PreparedIsCurrent(p *PreparedTransaction) bool {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	return p.current(context.Background()) == nil
}
func NewDetachedExecution(e *Execution) (*PreparedTransaction, error) { return Detached(e, nil) }

func ReceivingKeySnapshot(w *Wallet) *primitives.ReceivingKeys {
	if w.receivingKeys == nil {
		return nil
	}
	return &primitives.ReceivingKeys{ViewingPublicKey: w.ViewingPublicKey(), ClueKey: w.ClueKey()}
}
