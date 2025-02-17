package service

import (
	"context"
	model "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
)

func (s *Service) AcceptTerms(ctx context.Context, address, publicKey string) error {
	termsAcceptance := &model.TermsAcceptance{
		Address:   address,
		PublicKey: publicKey,
	}

	return s.DbClients.SharedDBClient.SaveTermsAcceptance(ctx, termsAcceptance)
}
