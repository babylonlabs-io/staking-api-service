package v2service

import (
	"errors"
	"testing"

	"github.com/babylonlabs-io/babylon-staking-indexer/testutil"
	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateCanExpand(t *testing.T) {
	ctx := t.Context()

	testHash1, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash2, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash3, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash4, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash5, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash6, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash7, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)

	tests := []struct {
		name                     string
		delegation               indexerdbmodel.IndexerDelegationDetails
		allowList                map[string]bool
		babylonParams            []*indexertypes.BbnStakingParams
		babylonParamsError       error
		expectedResult           bool
		expectedErrorInGetParams bool
	}{
		{
			name: "Active delegation with single FP, under max limit, in allow-list",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          testHash1,
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				testHash1: true,
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true,
		},
		{
			name: "Active delegation with max FPs, should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          testHash2,
				FinalityProviderBtcPksHex: []string{"fp1", "fp2", "fp3"},
			},
			allowList: map[string]bool{
				testHash2: true,
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 3},
			},
			expectedResult: false,
		},
		{
			name: "Inactive delegation should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateWithdrawn,
				StakingTxHashHex:          testHash3,
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				testHash3: true,
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false,
		},
		{
			name: "Active delegation not in allow-list should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          testHash4,
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"other-hash": true,
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false,
		},
		{
			name: "Active delegation with no allow-list configured should expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          testHash5,
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{}, // Empty allow-list
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true,
		},
		{
			name: "Multiple babylon params versions, should use latest",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          testHash6,
				FinalityProviderBtcPksHex: []string{"fp1", "fp2"},
			},
			allowList: map[string]bool{}, // No allow-list
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
				{Version: 3, MaxFinalityProviders: 2}, // Latest version with lower limit
				{Version: 2, MaxFinalityProviders: 10},
			},
			expectedResult: false, // Should use version 3 with MaxFinalityProviders=2
		},
		{
			name: "Error getting babylon params should return false",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          testHash7,
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList:                map[string]bool{},
			babylonParamsError:       errors.New("database error"),
			expectedResult:           false,
			expectedErrorInGetParams: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexerDB := mocks.NewIndexerDBClient(t)

			// Only mock GetBbnStakingParams call if delegation is active (will reach the params check)
			if tt.delegation.State == indexertypes.StateActive {
				if tt.expectedErrorInGetParams {
					indexerDB.On("GetBbnStakingParams", ctx).Return(nil, tt.babylonParamsError).Once()
				} else {
					indexerDB.On("GetBbnStakingParams", ctx).Return(tt.babylonParams, nil).Once()
				}
			}

			cfg := &config.Config{}
			dbClients := &dbclients.DbClients{
				IndexerDBClient: indexerDB,
			}

			sharedService, err := service.New(cfg, nil, nil, nil, dbClients)
			require.NoError(t, err)

			v2Service, err := New(sharedService, nil, tt.allowList)
			require.NoError(t, err)

			result := v2Service.evaluateCanExpand(ctx, tt.delegation)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetLatestMaxFinalityProviders(t *testing.T) {
	ctx := t.Context()

	tests := []struct {
		name           string
		babylonParams  []*indexertypes.BbnStakingParams
		paramsError    error
		expectedResult uint32
		expectError    bool
	}{
		{
			name: "Single params version",
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: 5,
			expectError:    false,
		},
		{
			name: "Multiple params versions, should return latest",
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 3},
				{Version: 3, MaxFinalityProviders: 7},
				{Version: 2, MaxFinalityProviders: 5},
			},
			expectedResult: 7, // Version 3 is the latest
			expectError:    false,
		},
		{
			name: "First element has highest MaxFinalityProviders but not highest version",
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 10}, // Highest MaxFinalityProviders but lowest version
				{Version: 5, MaxFinalityProviders: 3},  // Highest version
				{Version: 3, MaxFinalityProviders: 7},
			},
			expectedResult: 3, // Version 5 is the latest, should return its MaxFinalityProviders (3)
			expectError:    false,
		},
		{
			name:          "No params found",
			babylonParams: []*indexertypes.BbnStakingParams{},
			expectError:   true,
		},
		{
			name:        "Database error",
			paramsError: errors.New("database connection failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexerDB := mocks.NewIndexerDBClient(t)

			if tt.paramsError != nil {
				indexerDB.On("GetBbnStakingParams", ctx).Return(nil, tt.paramsError).Once()
			} else {
				indexerDB.On("GetBbnStakingParams", ctx).Return(tt.babylonParams, nil).Once()
			}

			cfg := &config.Config{}
			dbClients := &dbclients.DbClients{
				IndexerDBClient: indexerDB,
			}

			sharedService, err := service.New(cfg, nil, nil, nil, dbClients)
			require.NoError(t, err)

			v2Service, err := New(sharedService, nil, nil)
			require.NoError(t, err)

			result, err := v2Service.getLatestMaxFinalityProviders(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
