package v1service

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"github.com/rs/zerolog/log"
)

type FpDescriptionPublic struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

var emptyFpDescriptionPublic = &FpDescriptionPublic{
	Moniker:         "",
	Identity:        "",
	Website:         "",
	SecurityContact: "",
	Details:         "",
}

type FpDetailsPublic struct {
	Description       *FpDescriptionPublic `json:"description"`
	Commission        string               `json:"commission"`
	BtcPk             string               `json:"btc_pk"`
	ActiveTvl         int64                `json:"active_tvl"`
	TotalTvl          int64                `json:"total_tvl"`
	ActiveDelegations int64                `json:"active_delegations"`
	TotalDelegations  int64                `json:"total_delegations"`
}

type FpParamsPublic struct {
	Description *FpDescriptionPublic `json:"description"`
	Commission  string               `json:"commission"`
	BtcPk       string               `json:"btc_pk"`
}

// GetFinalityProvidersFromGlobalParams returns the finality providers from the global params.
// Those FP are treated as "active" finality providers.
func (s *V1Service) GetFinalityProvidersFromGlobalParams() []*FpParamsPublic {
	var fpDetails []*FpParamsPublic
	for _, finalityProvider := range s.Service.FinalityProviders {
		description := &FpDescriptionPublic{
			Moniker:         finalityProvider.Description.Moniker,
			Identity:        finalityProvider.Description.Identity,
			Website:         finalityProvider.Description.Website,
			SecurityContact: finalityProvider.Description.SecurityContact,
			Details:         finalityProvider.Description.Details,
		}
		fpDetails = append(fpDetails, &FpParamsPublic{
			Description: description,
			Commission:  finalityProvider.Commission,
			BtcPk:       finalityProvider.BtcPk,
		})
	}
	return fpDetails
}

func (s *V1Service) GetFinalityProvider(
	ctx context.Context, fpPkHex string,
) (*FpDetailsPublic, *types.Error) {
	fpStatsByPks, err :=
		s.Service.DbClients.V1DBClient.FindFinalityProviderStatsByFinalityProviderPkHex(
			ctx, []string{fpPkHex},
		)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Msg("Error while fetching finality provider from DB")
		return nil, types.NewInternalServiceError(err)
	}

	fpParams := s.GetFinalityProvidersFromGlobalParams()
	// Found nothing in the DB, we will try get the FP from global params
	if len(fpStatsByPks) == 0 {
		for _, fp := range fpParams {
			if fp.BtcPk == fpPkHex {
				return &FpDetailsPublic{
					Description:       fp.Description,
					Commission:        fp.Commission,
					BtcPk:             fp.BtcPk,
					ActiveTvl:         0,
					TotalTvl:          0,
					ActiveDelegations: 0,
					TotalDelegations:  0,
				}, nil
			}
		}
		// Return nil as nothing was found
		return nil, nil
	}
	if len(fpStatsByPks) > 1 {
		log.Ctx(ctx).Error().
			Str("fpPkHex", fpPkHex).
			Int("numFinalityProviders", len(fpStatsByPks)).
			Msg("Found more than one finality provider with the same pk")
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError, types.InternalServiceError,
			"Found more than one finality provider with the same pk",
		)
	}
	fpStat := fpStatsByPks[0]

	var fpParamsPublic *FpParamsPublic
	for _, fp := range fpParams {
		if fp.BtcPk == fpStat.FinalityProviderPkHex {
			fpParamsPublic = fp
			break
		}
	}
	if fpParamsPublic == nil {
		fpParamsPublic = &FpParamsPublic{
			Description: emptyFpDescriptionPublic,
			Commission:  "",
			BtcPk:       fpStat.FinalityProviderPkHex,
		}
	}
	return &FpDetailsPublic{
		Description:       fpParamsPublic.Description,
		Commission:        fpParamsPublic.Commission,
		BtcPk:             fpStat.FinalityProviderPkHex,
		ActiveTvl:         fpStat.ActiveTvl,
		TotalTvl:          fpStat.TotalTvl,
		ActiveDelegations: fpStat.ActiveDelegations,
		TotalDelegations:  fpStat.TotalDelegations,
	}, nil
}

func (s *V1Service) GetFinalityProviders(ctx context.Context, page string) ([]*FpDetailsPublic, string, *types.Error) {
	fpParams := s.GetFinalityProvidersFromGlobalParams()
	if len(fpParams) == 0 {
		log.Ctx(ctx).Error().Msg("No finality providers found from global params")
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "No finality providers found from global params")
	}
	// Convert the fpParams slice to a map with the BtcPk as the key
	fpParamsMap := make(map[string]*FpParamsPublic)
	for _, fp := range fpParams {
		fpParamsMap[fp.BtcPk] = fp
	}

	resultMap, err := s.Service.DbClients.V1DBClient.FindFinalityProviderStats(ctx, page)
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Invalid pagination token when fetching finality providers")
			return nil, "", types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		// We don't want to return an error here in case of DB error.
		// we will continue the process with the data we have from global params as a fallback.
		// TODO: Add metric for this error and alerting
		log.Ctx(ctx).Error().Err(err).Msg("Error while fetching finality providers from DB")
		// Return the finality providers from global params as a fallback
		return buildFallbackFpDetailsPublic(fpParams), "", nil
	}

	/*
		If no finality providers are found in the database and no pagination token
		is provided (indicating this is the first page), return the finality providers
		from the global parameters as a fallback. This fallback is only necessary when
		launching the service for the first time and no finality providers are found in the database.
	*/
	if (len(resultMap.Data) == 0) && (page == "") {
		return buildFallbackFpDetailsPublic(fpParams), "", nil
	}

	var finalityProviderDetailsPublic []*FpDetailsPublic
	for _, fp := range resultMap.Data {
		var paramsPublic *FpParamsPublic
		if fpParamsMap[fp.FinalityProviderPkHex] != nil {
			paramsPublic = fpParamsMap[fp.FinalityProviderPkHex]
		} else {
			paramsPublic = &FpParamsPublic{
				Description: emptyFpDescriptionPublic,
				Commission:  "",
				BtcPk:       fp.FinalityProviderPkHex,
			}
		}

		detail := &FpDetailsPublic{
			Description:       paramsPublic.Description,
			Commission:        paramsPublic.Commission,
			BtcPk:             fp.FinalityProviderPkHex,
			ActiveTvl:         fp.ActiveTvl,
			TotalTvl:          fp.TotalTvl,
			ActiveDelegations: fp.ActiveDelegations,
			TotalDelegations:  fp.TotalDelegations,
		}
		finalityProviderDetailsPublic = append(finalityProviderDetailsPublic, detail)
	}
	// If there are no more pages to fetch, make sure all the finality providers from global params are included
	if resultMap.PaginationToken == "" {
		fpsNotInUse, err := s.FindRegisteredFinalityProvidersNotInUse(ctx, fpParams)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Error while fetching finality providers not in use")
			return nil, "", types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
		}

		finalityProviderDetailsPublic = append(finalityProviderDetailsPublic, fpsNotInUse...)
	}

	return finalityProviderDetailsPublic, resultMap.PaginationToken, nil
}

func (s *V1Service) FindRegisteredFinalityProvidersNotInUse(
	ctx context.Context, fpParams []*FpParamsPublic,
) ([]*FpDetailsPublic, error) {
	var finalityProvidersPkHex []string
	for _, fp := range fpParams {
		finalityProvidersPkHex = append(finalityProvidersPkHex, fp.BtcPk)
	}
	fpStatsByPks, err := s.Service.DbClients.V1DBClient.FindFinalityProviderStatsByFinalityProviderPkHex(ctx, finalityProvidersPkHex)
	if err != nil {
		return nil, err
	}
	fpStatsByPksMap := make(map[string]*v1model.FinalityProviderStatsDocument)
	for _, fpStat := range fpStatsByPks {
		fpStatsByPksMap[fpStat.FinalityProviderPkHex] = fpStat
	}

	// Find the finality providers that are not in the fpStatsByPksMap
	var fps []*FpDetailsPublic
	for _, fp := range fpParams {
		if fpStatsByPksMap[fp.BtcPk] == nil {
			detail := &FpDetailsPublic{
				Description:       fp.Description,
				Commission:        fp.Commission,
				BtcPk:             fp.BtcPk,
				ActiveTvl:         0,
				TotalTvl:          0,
				ActiveDelegations: 0,
				TotalDelegations:  0,
			}
			fps = append(fps, detail)
		}
	}
	return fps, nil
}

func buildFallbackFpDetailsPublic(fpParams []*FpParamsPublic) []*FpDetailsPublic {
	var finalityProviderDetailsPublic []*FpDetailsPublic
	for _, fp := range fpParams {
		detail := &FpDetailsPublic{
			Description:       fp.Description,
			Commission:        fp.Commission,
			BtcPk:             fp.BtcPk,
			ActiveTvl:         0,
			TotalTvl:          0,
			ActiveDelegations: 0,
			TotalDelegations:  0,
		}
		finalityProviderDetailsPublic = append(finalityProviderDetailsPublic, detail)
	}
	return finalityProviderDetailsPublic
}
