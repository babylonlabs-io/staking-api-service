package v2service

import (
	"context"
	"math/rand"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
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
	// FinalityProviderDelegations int64                `json:"finality_provider_delegations,omitempty"`
}

type FinalityProvidersPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}

func (s *V2Service) GetFinalityProviders(ctx context.Context, paginationKey string) ([]FinalityProviderPublic, string, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// random number of providers between 1 and 10
	numProviders := testutils.RandomPositiveInt(r, 10)
	providers := testutils.GenerateRandomFinalityProviderDetail(r, uint64(numProviders))
	publicProviders := make([]FinalityProviderPublic, len(providers))
	for i, provider := range providers {
		publicProviders[i] = FinalityProviderPublic{
			BtcPK:             testutils.GeneratePks(1)[0],
			State:             testutils.RandomFinalityProviderState(r),
			Description:       provider.Description,
			Commission:        provider.Commission,
			ActiveTVL:         int64(testutils.RandomPositiveInt(r, 1000000000000000000)),
			TotalTVL:          int64(testutils.RandomPositiveInt(r, 1000000000000000000)),
			ActiveDelegations: int64(testutils.RandomPositiveInt(r, 100)),
			TotalDelegations:  int64(testutils.RandomPositiveInt(r, 100)),
		}
	}

	return publicProviders, "", nil
}
