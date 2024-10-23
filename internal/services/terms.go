package services

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

func (s *Services) AcceptTerms(ctx context.Context, address, publicKey string, termsAccepted bool) *types.Error {
	termsAcceptance := &model.TermsAcceptance{
		Address:       address,
		PublicKey:     publicKey,
		TermsAccepted: termsAccepted,
	}

	if err := s.DbClient.SaveTermsAcceptance(ctx, termsAcceptance); err != nil {
		return types.NewInternalServiceError(err)
	}

	return nil
}
