package v2service

import (
	"context"
	"math/rand"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
)

type OverallStatsPublic struct {
	ActiveTVL               int64 `json:"active_tvl"`
	TotalTVL                int64 `json:"total_tvl"`
	ActiveDelegations       int64 `json:"active_delegations"`
	TotalDelegations        int64 `json:"total_delegations"`
	ActiveStakers           int64 `json:"active_stakers"`
	TotalStakers            int64 `json:"total_stakers"`
	ActiveFinalityProviders int64 `json:"active_finality_providers"`
	TotalFinalityProviders  int64 `json:"total_finality_providers"`
}

func (s *V2Service) GetOverallStats(ctx context.Context) (OverallStatsPublic, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	overallStats := OverallStatsPublic{
		ActiveTVL:               int64(testutils.RandomPositiveInt(r, 1000000)),
		TotalTVL:                int64(testutils.RandomPositiveInt(r, 1000000)),
		ActiveDelegations:       int64(testutils.RandomPositiveInt(r, 100)),
		TotalDelegations:        int64(testutils.RandomPositiveInt(r, 100)),
		ActiveStakers:           int64(testutils.RandomPositiveInt(r, 100)),
		TotalStakers:            int64(testutils.RandomPositiveInt(r, 100)),
		ActiveFinalityProviders: int64(testutils.RandomPositiveInt(r, 100)),
		TotalFinalityProviders:  int64(testutils.RandomPositiveInt(r, 100)),
	}
	return overallStats, nil
}
