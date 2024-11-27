package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type FinalityProviderStatsPublic struct {
	BtcPk             string                              `json:"btc_pk"`
	State             types.FinalityProviderQueryingState `json:"state"`
	Description       types.FinalityProviderDescription   `json:"description"`
	Commission        string                              `json:"commission"`
	ActiveTvl         int64                               `json:"active_tvl"`
	ActiveDelegations int64                               `json:"active_delegations"`
}

type FinalityProvidersStatsPublic struct {
	FinalityProviders []FinalityProviderStatsPublic `json:"finality_providers"`
}

func mapToFinalityProviderStatsPublic(provider indexerdbmodel.IndexerFinalityProviderDetails) *FinalityProviderStatsPublic {
	return &FinalityProviderStatsPublic{
		BtcPk:             provider.BtcPk,
		State:             types.FinalityProviderQueryingState(provider.State),
		Description:       types.FinalityProviderDescription(provider.Description),
		Commission:        provider.Commission,
		ActiveTvl:         0,
		ActiveDelegations: 0,
	}
}

// GetFinalityProviders gets a list of finality providers with stats
func (s *V2Service) GetFinalityProvidersWithStats(
	ctx context.Context,
) ([]*FinalityProviderStatsPublic, *types.Error) {
	fps, err := s.DbClients.IndexerDBClient.GetFinalityProviders(ctx)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Finality providers not found")
			return nil, types.NewErrorWithMsg(
				http.StatusNotFound, types.NotFound, "finality providers not found, please retry",
			)
		}
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError, types.InternalServiceError, "failed to get finality providers",
		)
	}

	// TODO: Call the FP stats service to get the stats for compose the response
	providersPublic := make([]*FinalityProviderStatsPublic, 0, len(fps))

	for _, provider := range fps {
		providersPublic = append(providersPublic, mapToFinalityProviderStatsPublic(*provider))
	}
	return providersPublic, nil
}
