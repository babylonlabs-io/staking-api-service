package services

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

func (s *Services) AcceptTerms(ctx context.Context, address, publicKey string) *types.Error {
	termsAcceptance := &model.TermsAcceptance{
		Address:   address,
		PublicKey: publicKey,
	}

	if err := s.DbClient.SaveTermsAcceptance(ctx, termsAcceptance); err != nil {
		return types.NewInternalServiceError(err)
	}

	return nil
}
