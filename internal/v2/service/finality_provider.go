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
	BsnID             string                              `json:"bsn_id,omitempty"`
	Type              string                              `json:"type"`
}

type FinalityProvidersStatsPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}

func mapToFinalityProviderStatsPublic(
	provider indexerdbmodel.IndexerFinalityProviderDetails,
	fpStats *v2dbmodel.V2FinalityProviderStatsDocument,
	bsn *indexerdbmodel.BSN,
	fpLogoURL string,
) *FinalityProviderPublic {
	var bsnType string
	if bsn != nil {
		switch bsn.Type {
		case indexerdbmodel.TypeCosmos:
			bsnType = "cosmos"
		case indexerdbmodel.TypeRollup:
			bsnType = "rollup"
		}
	}

	return &FinalityProviderPublic{
		BtcPk:             provider.BtcPk,
		State:             types.FinalityProviderQueryingState(provider.State),
		Description:       types.FinalityProviderDescription(provider.Description),
		Commission:        provider.Commission,
		ActiveTvl:         fpStats.ActiveTvl,
		ActiveDelegations: fpStats.ActiveDelegations,
		LogoURL:           fpLogoURL,
		BsnID:             provider.BsnID,
		Type:              bsnType,
	}
}

// GetFinalityProvidersWithStats retrieves all finality providers and their associated statistics
func (s *V2Service) GetFinalityProvidersWithStats(
	ctx context.Context,
	bsnID *string,
) ([]*FinalityProviderPublic, *types.Error) {
	if bsnID == nil {
		// if no bsn_id is provided we first retrieve chain_id corresponding to babylon network
		// then we filter all finality providers by bsn_id = chain_id so we end up with default behavior:
		// in response there will be only finality providers for babylon
		networkInfo, err := s.dbClients.IndexerDBClient.GetNetworkInfo(ctx)
		if err != nil {
			return nil, types.NewErrorWithMsg(
				http.StatusInternalServerError,
				types.InternalServiceError,
				"failed to get network info",
			)
		}

		bsnID = &networkInfo.ChainID
	}

	finalityProviders, err := s.dbClients.IndexerDBClient.GetFinalityProviders(ctx, bsnID)
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

	logoMap := s.fetchLogos(ctx, finalityProviders)

	statsLookup := make(map[string]*v2dbmodel.V2FinalityProviderStatsDocument)
	for _, stats := range providerStats {
		statsLookup[stats.FinalityProviderPkHex] = stats
	}

	finalityProvidersPublic := make([]*FinalityProviderPublic, 0, len(finalityProviders))

	bsn, err := s.dbClients.IndexerDBClient.GetAllBSN(ctx)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get bsn list",
		)
	}
	bsnMap := pkg.SliceToMap(bsn, func(b indexerdbmodel.BSN) string {
		return b.ID
	})

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

		var logoURL string
		if logoMap != nil {
			logoURL = logoMap[provider.BtcPk]
		}

		var bsn *indexerdbmodel.BSN
		if bsnValue, ok := bsnMap[provider.BsnID]; ok {
			bsn = &bsnValue
		}
		finalityProvidersPublic = append(
			finalityProvidersPublic,
			mapToFinalityProviderStatsPublic(*provider, providerStats, bsn, logoURL),
		)
	}
	return finalityProvidersPublic, nil
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
