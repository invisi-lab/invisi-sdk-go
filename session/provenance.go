package session

import (
	"context"

	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func (o *Operation) ExportProvenanceTracking(ctx context.Context) ([]provenance.Tracking, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution := o.execution()
	return execution.ExportProvenanceTracking(ctx)
}

func (o *Operation) ProveProvenance(ctx context.Context, options ...sdk.ProvenanceProveOptions) (*Operation, error) {
	return o.stage(ctx, func(next *Operation) error {
		execution := next.execution()
		if e := execution.ProveProvenance(ctx, options...); e != nil {
			return e
		}
		next.plan = execution.Plan
		return nil
	})
}
func (o *Operation) ExportProvenanceProof(ctx context.Context) (provenance.ProvenanceProof, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return provenance.ProvenanceProof{}, e
	}
	return o.execution().ExportProvenanceProof(ctx)
}
func (o *Operation) SubmitProvenance(ctx context.Context) (provenance.Result, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return nil, e
	}
	return o.execution().SubmitProvenance(ctx)
}
