package v2service

import (
	"errors"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainanalysis"
)

func (s *V2Service) AssessAddress(address string) (*chainanalysis.AddressAssessment, error) {
	// only possible if corresponding config is empty
	if s.Clients.ChainAnalysis == nil {
		return nil, errors.New("ChainAnalysis client is not initialized")
	}

	return s.Clients.ChainAnalysis.AssessAddress(address)
}
