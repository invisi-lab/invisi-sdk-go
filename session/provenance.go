package session

import (
	"context"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
)

func (o *Operation) ExportProvenanceTracking(ctx context.Context) ([]provenance.Tracking, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	return execution.ExportProvenanceTracking(ctx)
}
