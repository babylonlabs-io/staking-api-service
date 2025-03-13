package v2service

import (
	"context"
	"errors"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainalysis"
)

func (s *V2Service) AssessAddress(ctx context.Context, address string) (*chainalysis.AddressAssessment, error) {
	// only possible if corresponding config is empty
	if s.Clients.Chainalysis == nil {
		return nil, errors.New("Chainalysis client is not initialized")
	}

	return s.Clients.Chainalysis.AssessAddress(ctx, address)
}
