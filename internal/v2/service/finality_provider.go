package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"github.com/rs/zerolog/log"
)

type FinalityProviderStatsPublic struct {
	BtcPk              string                               `json:"btc_pk"`
	State              indexerdbmodel.FinalityProviderState `json:"state"`
	Description        types.FinalityProviderDescription    `json:"description"`
	Commission         string                               `json:"commission"`
	ActiveTvl          int64                                `json:"active_tvl"`
	ActiveDelegations  int64                                `json:"active_delegations"`
	TransitionRequired bool                                 `json:"transition_required"`
}

type FinalityProvidersStatsPublic struct {
	FinalityProviders []FinalityProviderStatsPublic `json:"finality_providers"`
}

func mapIndexerFpToFinalityProviderStatsPublic(provider indexerdbmodel.IndexerFinalityProviderDetails) *FinalityProviderStatsPublic {
	return &FinalityProviderStatsPublic{
		BtcPk:              provider.BtcPk,
		State:              provider.State,
		Description:        types.FinalityProviderDescription(provider.Description),
		Commission:         provider.Commission,
		ActiveTvl:          0,
		ActiveDelegations:  0,
		TransitionRequired: false,
	}
}

// mapV1FpStatsToFinalityProviderStatsPublic maps a V1 finality provider to a public finality provider stats
func mapV1FpStatsToFinalityProviderStatsPublic(provider types.FinalityProviderDetails) *FinalityProviderStatsPublic {
	return &FinalityProviderStatsPublic{
		BtcPk:              provider.BtcPk,
		State:              indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE, // v1 only has active FPs
		Description:        provider.Description,
		Commission:         provider.Commission,
		ActiveTvl:          0,
		ActiveDelegations:  0,
		TransitionRequired: false,
	}
}

// GetFinalityProviders gets a list of finality providers with stats
func (s *V2Service) GetFinalityProvidersWithStats(
	ctx context.Context,
) ([]*FinalityProviderStatsPublic, *types.Error) {

	// For V1
	v1FinalityProviders, err := s.getFinalityProvidersFromGlobalParamsWithStats(ctx)
	if err != nil {
		// Handle the error appropriately
		return nil, err
	}
	v1FinalityProvidersMap := make(map[string]*types.FinalityProviderDescription)
	for _, fp := range v1FinalityProviders {
		v1FinalityProvidersMap[fp.BtcPk] = &fp.Description
	}

	fps, dbErr := s.DbClients.IndexerDBClient.GetFinalityProviders(ctx)
	if dbErr != nil {
		if db.IsNotFoundError(dbErr) {
			log.Ctx(ctx).Warn().Err(dbErr).Msg("Finality providers not found")
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
		providersPublic = append(providersPublic, mapIndexerFpToFinalityProviderStatsPublic(*provider))
	}

	// Create a map of V2 providers for lookup
	v2ProvidersMap := make(map[string]bool)
	for _, provider := range fps {
		v2ProvidersMap[provider.BtcPk] = true
	}

	// Add V1 providers that aren't in V2, marking them as requiring transition
	for _, v1Provider := range v1FinalityProviders {
		if !v2ProvidersMap[v1Provider.BtcPk] {
			v1Provider.TransitionRequired = true
			providersPublic = append(providersPublic, v1Provider)
		}
	}

	// Return the combined list of V1 and V2 finality providers
	return providersPublic, nil
}

// For V1: getFinalityProvidersFromGlobalParams returns the finality providers from the global params.
// Those FP are treated as "active" finality providers.
func (s *V2Service) getFinalityProvidersFromGlobalParamsWithStats(ctx context.Context) ([]*FinalityProviderStatsPublic, *types.Error) {
	var fpDetails []*FinalityProviderStatsPublic
	for _, finalityProvider := range s.FinalityProviders {
		fpDetails = append(fpDetails, mapV1FpStatsToFinalityProviderStatsPublic(finalityProvider))
	}

	// Get the stats for the finality providers, page token is empty as we are now fetching all the finality providers on first page
	resultMap, err := s.DbClients.V1DBClient.FindFinalityProviderStats(ctx, "")
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Invalid pagination token when fetching finality providers")
			return nil, types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		log.Ctx(ctx).Error().Err(err).Msg("Error while fetching finality providers from DB")
		return fpDetails, nil
	}

	if len(resultMap.Data) == 0 {
		return fpDetails, nil
	}

	fpDetailsMap := make(map[string]*FinalityProviderStatsPublic)
	for _, fp := range fpDetails {
		fpDetailsMap[fp.BtcPk] = fp
	}

	for _, fp := range resultMap.Data {
		var paramsPublic *FinalityProviderStatsPublic
		if fpDetailsMap[fp.FinalityProviderPkHex] != nil {
			paramsPublic = fpDetailsMap[fp.FinalityProviderPkHex]
		} else {
			paramsPublic = &FinalityProviderStatsPublic{
				Description: types.FinalityProviderDescription{},
				Commission:  "",
				BtcPk:       fp.FinalityProviderPkHex,
			}
		}

		detail := &FinalityProviderStatsPublic{
			Description:       paramsPublic.Description,
			State:             indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE,
			Commission:        paramsPublic.Commission,
			BtcPk:             fp.FinalityProviderPkHex,
			ActiveTvl:         fp.ActiveTvl,
			ActiveDelegations: fp.ActiveDelegations,
		}
		fpDetails = append(fpDetails, detail)
	}

	// Make sure all the finality providers from global params are included
	if resultMap.PaginationToken == "" {
		fpsNotInUse, err := s.findRegisteredFinalityProvidersNotInUse(ctx, fpDetails)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Error while fetching finality providers not in use")
			return nil, types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
		}

		fpDetails = append(fpDetails, fpsNotInUse...)
	}

	return fpDetails, nil
}

// For V1: FindRegisteredFinalityProvidersNotInUse finds the registered finality providers that are not in use
func (s *V2Service) findRegisteredFinalityProvidersNotInUse(
	ctx context.Context, fpParams []*FinalityProviderStatsPublic,
) ([]*FinalityProviderStatsPublic, *types.Error) {
	var finalityProvidersPkHex []string
	for _, fp := range fpParams {
		finalityProvidersPkHex = append(finalityProvidersPkHex, fp.BtcPk)
	}
	fpStatsByPks, err := s.DbClients.V1DBClient.FindFinalityProviderStatsByFinalityProviderPkHex(ctx, finalityProvidersPkHex)
	if err != nil {
		return nil, types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}
	fpStatsByPksMap := make(map[string]*v1model.FinalityProviderStatsDocument)
	for _, fpStat := range fpStatsByPks {
		fpStatsByPksMap[fpStat.FinalityProviderPkHex] = fpStat
	}

	// Find the finality providers that are not in the fpStatsByPksMap
	var fps []*FinalityProviderStatsPublic
	for _, fp := range fpParams {
		if fpStatsByPksMap[fp.BtcPk] == nil {
			detail := &FinalityProviderStatsPublic{
				Description:        fp.Description,
				State:              indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE,
				Commission:         fp.Commission,
				BtcPk:              fp.BtcPk,
				ActiveTvl:          0,
				ActiveDelegations:  0,
				TransitionRequired: false,
			}
			fps = append(fps, detail)
		}
	}
	return fps, nil
}
