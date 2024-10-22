package v2service

import (
	"context"
	"math/rand"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils/datagen"
)

type FinalityProviderPublic struct {
	BtcPK             string                            `json:"btc_pk"`
	State             types.FinalityProviderState       `json:"state"`
	Description       types.FinalityProviderDescription `json:"description"`
	Commission        string                            `json:"commission"`
	ActiveTVL         int64                             `json:"active_tvl"`
	TotalTVL          int64                             `json:"total_tvl"`
	ActiveDelegations int64                             `json:"active_delegations"`
	TotalDelegations  int64                             `json:"total_delegations"`
}

type FinalityProvidersPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}

func (s *V2Service) GetFinalityProviders(ctx context.Context, paginationKey string) ([]FinalityProviderPublic, string, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// random number of providers between 1 and 10
	numProviders := datagen.RandomPositiveInt(r, 10)
	providers := datagen.GenerateRandomFinalityProviderDetail(r, uint64(numProviders))
	publicProviders := make([]FinalityProviderPublic, len(providers))
	for i, provider := range providers {
		publicProviders[i] = FinalityProviderPublic{
			BtcPK:             datagen.GeneratePks(1)[0],
			State:             datagen.RandomFinalityProviderState(r),
			Description:       provider.Description,
			Commission:        provider.Commission,
			ActiveTVL:         int64(datagen.RandomPositiveInt(r, 1000000000000000000)),
			TotalTVL:          int64(datagen.RandomPositiveInt(r, 1000000000000000000)),
			ActiveDelegations: int64(datagen.RandomPositiveInt(r, 100)),
			TotalDelegations:  int64(datagen.RandomPositiveInt(r, 100)),
		}
	}

	return publicProviders, "", nil
}
