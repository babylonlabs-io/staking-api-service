package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"github.com/rs/zerolog/log"
)

type FinalityProviderPublic struct {
	BtcPk             string                              `json:"btc_pk"`
	State             types.FinalityProviderQueryingState `json:"state"`
	Description       types.FinalityProviderDescription   `json:"description"`
	Commission        string                              `json:"commission"`
	ActiveTvl         int64                               `json:"active_tvl"`
	ActiveDelegations int64                               `json:"active_delegations"`
	LogoURL           string                              `json:"logo_url,omitempty"`
	Type              string                              `json:"type"`
}

type FinalityProvidersStatsPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}

func mapToFinalityProviderStatsPublic(
	provider indexerdbmodel.IndexerFinalityProviderDetails,
	fpStats *v2dbmodel.V2FinalityProviderStatsDocument,
	fpLogoURL string,
) *FinalityProviderPublic {
	return &FinalityProviderPublic{
		BtcPk:             provider.BtcPk,
		State:             types.FinalityProviderQueryingState(provider.State),
		Description:       types.FinalityProviderDescription(provider.Description),
		Commission:        provider.Commission,
		ActiveTvl:         fpStats.ActiveTvl,
		ActiveDelegations: fpStats.ActiveDelegations,
		LogoURL:           fpLogoURL,
		Type:              "",
	}
}

// GetFinalityProvidersWithStats retrieves finality providers sorted by active TVL with pagination
// Implementation follows stats-first approach: query stats sorted by active_tvl, then fetch details
func (s *V2Service) GetFinalityProvidersWithStats(
	ctx context.Context,
	paginationToken string,
) ([]*FinalityProviderPublic, string, *types.Error) {
	fpStatsResult, err := s.dbClients.IndexerDBClient.GetFinalityProviderStatsPaginated(ctx, paginationToken)
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Invalid pagination token when fetching finality provider stats")
			return nil, "", types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		return nil, "", types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get finality provider stats",
		)
	}

	fpStats := fpStatsResult.Data

	if len(fpStats) == 0 {
		log.Ctx(ctx).Warn().Msg("No finality provider stats found")
		return nil, "", types.NewErrorWithMsg(
			http.StatusNotFound,
			types.NotFound,
			"finality providers not found, please retry",
		)
	}

	fpPkHexes := make([]string, 0, len(fpStats))
	for _, stat := range fpStats {
		fpPkHexes = append(fpPkHexes, stat.FpBtcPkHex)
	}

	finalityProviders, err := s.dbClients.IndexerDBClient.GetFinalityProvidersByPks(ctx, fpPkHexes)
	if err != nil {
		return nil, "", types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get finality provider details",
		)
	}

	logoMap := s.fetchLogos(ctx, finalityProviders)

	detailsLookup := make(map[string]*indexerdbmodel.IndexerFinalityProviderDetails)
	for _, fp := range finalityProviders {
		detailsLookup[fp.BtcPk] = fp
	}

	finalityProvidersPublic := make([]*FinalityProviderPublic, 0, len(fpStats))

	for _, stat := range fpStats {
		fpDetails, hasDetails := detailsLookup[stat.FpBtcPkHex]
		if !hasDetails {
			log.Ctx(ctx).Debug().
				Str("finality_provider_pk_hex", stat.FpBtcPkHex).
				Msg("Finality provider has stats but no details, skipping")
			continue
		}

		v2Stats := &v2dbmodel.V2FinalityProviderStatsDocument{
			FinalityProviderPkHex: stat.FpBtcPkHex,
			ActiveTvl:             int64(stat.ActiveTvl),
			ActiveDelegations:     int64(stat.ActiveDelegations),
		}

		var logoURL string
		if logoMap != nil {
			logoURL = logoMap[fpDetails.BtcPk]
		}

		finalityProvidersPublic = append(
			finalityProvidersPublic,
			mapToFinalityProviderStatsPublic(*fpDetails, v2Stats, logoURL),
		)
	}

	return finalityProvidersPublic, fpStatsResult.PaginationToken, nil
}

func (s *V2Service) fetchLogos(ctx context.Context, fps []*indexerdbmodel.IndexerFinalityProviderDetails) map[string]string {
	log := log.Ctx(ctx)

	ids := pkg.Map(fps, func(v *indexerdbmodel.IndexerFinalityProviderDetails) string {
		return v.BtcPk
	})
	logos, err := s.dbClients.V2DBClient.GetFinalityProviderLogosByID(ctx, ids)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch logos")
		return nil
	}

	// btc pk => url
	logoMap := make(map[string]*string)
	for _, logo := range logos {
		logoMap[logo.Id] = logo.URL
	}

	type logoToUpdate struct {
		identity string
		btcPK    string
	}
	missingLogos := make(chan logoToUpdate, len(fps)) // upper bound for logos is len(fps)
	for _, fp := range fps {
		_, ok := logoMap[fp.BtcPk]
		if ok {
			continue
		}

		// identity used as id for logo retrieval
		if fp.Description.Identity == "" {
			continue
		}

		missingLogos <- logoToUpdate{
			identity: fp.Description.Identity,
			btcPK:    fp.BtcPk,
		}
	}
	close(missingLogos)

	go func() {
		for missingLogo := range missingLogos {
			// because this goroutine may take longer than the current request to our endpoint,
			// we need to use different context; otherwise all requests will be canceled
			fetchCtx := context.Background()
			url, err := s.keybaseClient.GetLogoURL(fetchCtx, missingLogo.identity)
			if err != nil {
				log.Err(err).Str("identity", missingLogo.identity).Msg("Failed to fetch logo")
			}

			// we store null in case url is empty string so we don't fetch failed logos every time
			var urlValue *string
			if url != "" {
				urlValue = &url
			}
			err = s.dbClients.V2DBClient.InsertFinalityProviderLogo(fetchCtx, missingLogo.btcPK, urlValue)
			if err != nil {
				log.Err(err).Str("identity", missingLogo.identity).Msg("Failed to insert logo url")
			}
		}
	}()

	result := make(map[string]string, len(logoMap))
	for id, url := range logoMap {
		if url == nil {
			continue
		}

		result[id] = *url
	}
	return result
}
