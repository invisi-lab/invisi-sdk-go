package session

import (
	"context"
	"fmt"

	"github.com/invisi-lab/invisi-sdk-go/credit"
)

func (s *Session) prepareRequest(ctx context.Context, id string, request any) (result *Operation, failure error) {
	defer func() {
		if s.wrapPreparation != nil {
			failure = s.wrapPreparation("prepareCredit", failure)
		}
	}()
	switch p := request.(type) {
	case credit.CreatePreparation:
		request = p.Clone()
	case credit.UpdatePreparation:
		request = p.Clone()
	case credit.DrawPreparation:
		request = p.Clone()
	case credit.RevokePreparation:
		request = p.Clone()
	}
	if s.build == nil {
		return nil, fmt.Errorf("Credit request preparation capability required")
	}
	prepared, err := s.build(ctx, request)
	if err != nil {
		return nil, err
	}
	return s.Prepare(ctx, id, prepared)
}
func (s *Session) PrepareCreate(ctx context.Context, id string, p credit.CreatePreparation) (*Operation, error) {
	return s.prepareRequest(ctx, id, p)
}
func (s *Session) PrepareUpdate(ctx context.Context, id string, p credit.UpdatePreparation) (*Operation, error) {
	return s.prepareRequest(ctx, id, p)
}
func (s *Session) PrepareDraw(ctx context.Context, id string, p credit.DrawPreparation) (*Operation, error) {
	return s.prepareRequest(ctx, id, p)
}
func (s *Session) PrepareRevoke(ctx context.Context, id string, p credit.RevokePreparation) (*Operation, error) {
	return s.prepareRequest(ctx, id, p)
}
