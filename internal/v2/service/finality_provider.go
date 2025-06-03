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
	}
}

// GetFinalityProvidersWithStats retrieves all finality providers and their associated statistics
func (s *V2Service) GetFinalityProvidersWithStats(
	ctx context.Context,
) ([]*FinalityProviderPublic, *types.Error) {
	finalityProviders, err := s.dbClients.IndexerDBClient.GetFinalityProviders(ctx)
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

	providerStats, err := s.dbClients.V2DBClient.GetFinalityProviderStats(ctx)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get finality provider stats",
		)
	}

	logoMap, err := s.fetchLogos(ctx, finalityProviders)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to get finality provider logos")
		// todo should we return an error here?
	}

	statsLookup := make(map[string]*v2dbmodel.V2FinalityProviderStatsDocument)
	for _, stats := range providerStats {
		statsLookup[stats.FinalityProviderPkHex] = stats
	}

	finalityProvidersPublic := make([]*FinalityProviderPublic, 0, len(finalityProviders))

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
		logoURL := logoMap[provider.BtcPk]

		finalityProvidersPublic = append(
			finalityProvidersPublic,
			mapToFinalityProviderStatsPublic(*provider, providerStats, logoURL),
		)
	}
	return finalityProvidersPublic, nil
}

func (s *V2Service) fetchLogos(ctx context.Context, fps []*indexerdbmodel.IndexerFinalityProviderDetails) (map[string]string, error) {
	ids := pkg.Map(fps, func(v *indexerdbmodel.IndexerFinalityProviderDetails) string {
		return v.BtcPk
	})
	logos, err := s.dbClients.V2DBClient.GetFinalityProviderLogosByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	// btc pk => url
	logoMap := make(map[string]string)
	for _, logo := range logos {
		logoMap[logo.Id] = logo.URL
	}

	// btc pk => identity
	missingLogos := make(map[string]string)
	for _, fp := range fps {
		_, ok := logoMap[fp.BtcPk]
		if ok {
			continue
		}

		missingLogos[fp.BtcPk] = fp.Description.Identity
	}

	log := log.Ctx(ctx)

	for btcPK, identity := range missingLogos {
		go func() {
			// todo add singleflight
			url, err := s.keybaseClient.GetLogoURL(ctx, identity)
			if err != nil {
				log.Err(err).Str("identity", identity).Msg("Failed to get logo url")
				return
			}

			err = s.dbClients.V2DBClient.InsertFinalityProviderLogo(ctx, btcPK, url)
			if err != nil {
				log.Err(err).Str("identity", identity).Msg("Failed to insert logo url")
			}
		}()
	}

	return logoMap, nil
}
