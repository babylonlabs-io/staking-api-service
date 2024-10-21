package v2service

import (
	"context"
	"math/rand"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
)

type StakerDelegationPublic struct {
	StakingTxHashHex      string                `json:"staking_tx_hash_hex"`
	StakerPKHex           string                `json:"staker_pk_hex"`
	FinalityProviderPKHex string                `json:"finality_provider_pk_hex"`
	StakingStartHeight    int64                 `json:"staking_start_height"`
	UnbondingStartHeight  int64                 `json:"unbonding_start_height"`
	Timelock              int64                 `json:"timelock"`
	StakingValue          int64                 `json:"staking_value"`
	State                 string                `json:"state"`
	StakingTx             types.TransactionInfo `json:"staking_tx"`
	UnbondingTx           types.TransactionInfo `json:"unbonding_tx"`
}

type StakerStatsPublic struct {
	StakerPKHex       string `json:"staker_pk_hex"`
	ActiveTVL         int64  `json:"active_tvl"`
	TotalTVL          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
}

func (s *V2Service) GetStakerDelegations(ctx context.Context, paginationKey string) ([]StakerDelegationPublic, string, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// random positive int
	numStakerDelegations := testutils.RandomPositiveInt(r, 10)
	stakerDelegationPublics := []StakerDelegationPublic{}
	for i := 0; i < numStakerDelegations; i++ {
		_, stakingTxHash, _ := testutils.GenerateRandomTx(r, nil)
		stakerPkHex, _ := testutils.RandomPk()
		fpPkHex, _ := testutils.RandomPk()
		stakerDelegation := &StakerDelegationPublic{
			StakingTxHashHex:      stakingTxHash,
			StakerPKHex:           stakerPkHex,
			FinalityProviderPKHex: fpPkHex,
			StakingStartHeight:    int64(testutils.RandomPositiveInt(r, 1000000)),
			UnbondingStartHeight:  int64(testutils.RandomPositiveInt(r, 1000000)),
			Timelock:              int64(testutils.RandomPositiveInt(r, 1000000)),
			StakingValue:          testutils.RandomAmount(r),
			State:                 types.Active.ToString(),
			StakingTx:             testutils.RandomTransactionInfo(r),
			UnbondingTx:           testutils.RandomTransactionInfo(r),
		}
		stakerDelegationPublics = append(stakerDelegationPublics, *stakerDelegation)
	}
	return stakerDelegationPublics, "", nil
}

func (s *V2Service) GetStakerStats(ctx context.Context, stakerPKHex string) (StakerStatsPublic, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	stakerStats := StakerStatsPublic{
		StakerPKHex:       stakerPKHex,
		ActiveTVL:         int64(testutils.RandomPositiveInt(r, 1000000)),
		TotalTVL:          int64(testutils.RandomPositiveInt(r, 1000000)),
		ActiveDelegations: int64(testutils.RandomPositiveInt(r, 100)),
		TotalDelegations:  int64(testutils.RandomPositiveInt(r, 100)),
	}
	return stakerStats, nil
}
