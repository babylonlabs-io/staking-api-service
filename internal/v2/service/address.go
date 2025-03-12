package v2service

import (
	"context"
	"errors"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainalysis"
)

func (s *V2Service) AssessAddress(ctx context.Context, address string) (*chainalysis.AddressAssessment, error) {
	// only possible if corresponding config is empty
	if s.Clients.ChainAnalysis == nil {
		return nil, errors.New("ChainAnalysis client is not initialized")
	}

	return s.Clients.ChainAnalysis.AssessAddress(ctx, address)
}
