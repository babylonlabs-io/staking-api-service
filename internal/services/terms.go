package services

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

func (s *Services) AcceptTerms(ctx context.Context, address, publicKey string, termsAccepted bool) *types.Error {
	if address == "" || publicKey == "" {
		return types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "Address and public key are required")
	}

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
