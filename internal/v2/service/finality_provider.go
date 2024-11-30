package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
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

func mapToFinalityProviderStatsPublic(
	provider indexerdbmodel.IndexerFinalityProviderDetails,
	fpStats *v2dbmodel.V2FinalityProviderStatsDocument,
) *FinalityProviderStatsPublic {
	return &FinalityProviderStatsPublic{
		BtcPk:             provider.BtcPk,
		State:             types.FinalityProviderQueryingState(provider.State),
		Description:       types.FinalityProviderDescription(provider.Description),
		Commission:        provider.Commission,
		ActiveTvl:         fpStats.ActiveTvl,
		ActiveDelegations: fpStats.ActiveDelegations,
	}
}

// GetFinalityProvidersWithStats retrieves all finality providers and their associated statistics
func (s *V2Service) GetFinalityProvidersWithStats(
	ctx context.Context,
) ([]*FinalityProviderStatsPublic, *types.Error) {
	finalityProviders, err := s.DbClients.IndexerDBClient.GetFinalityProviders(ctx)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("No finality providers found")
			return nil, types.NewErrorWithMsg(
				http.StatusNotFound,
				types.NotFound,
				"finality providers not found, please retry",
			)
		}
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get finality providers",
		)
	}

	providerStats, err := s.DbClients.V2DBClient.GetFinalityProviderStats(ctx)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get finality provider stats",
		)
	}

	statsLookup := make(map[string]*v2dbmodel.V2FinalityProviderStatsDocument)
	for _, stats := range providerStats {
		statsLookup[stats.FinalityProviderPkHex] = stats
	}

	finalityProvidersWithStats := make([]*FinalityProviderStatsPublic, 0, len(finalityProviders))

	for _, provider := range finalityProviders {
		providerStats, hasStats := statsLookup[provider.BtcPk]
		if !hasStats {
			providerStats = &v2dbmodel.V2FinalityProviderStatsDocument{
				ActiveTvl:         0,
				ActiveDelegations: 0,
			}
			log.Ctx(ctx).Debug().
				Str("finality_provider_pk_hex", provider.BtcPk).
				Msg("Initializing finality provider with default stats")
		}
		finalityProvidersWithStats = append(
			finalityProvidersWithStats,
			mapToFinalityProviderStatsPublic(*provider, providerStats),
		)
	}
	return finalityProvidersWithStats, nil
}
