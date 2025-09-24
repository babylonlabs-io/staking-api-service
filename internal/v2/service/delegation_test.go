package v2service

import (
	"errors"
	"testing"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2types "github.com/babylonlabs-io/staking-api-service/internal/v2/types"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDelegationState(t *testing.T) {
	tests := []struct {
		name           string
		state          indexertypes.DelegationState
		subState       indexertypes.DelegationSubState
		expectedResult v2types.DelegationState
		expectError    bool
	}{
		{
			name:           "Map StateExpanded with SubStateEarlyUnbonding",
			state:          indexertypes.StateExpanded,
			subState:       indexertypes.SubStateEarlyUnbonding,
			expectedResult: v2types.StateExpanded,
			expectError:    false,
		},
		{
			name:           "Map StateExpanded with SubStateTimelock",
			state:          indexertypes.StateExpanded,
			subState:       indexertypes.SubStateTimelock,
			expectedResult: v2types.StateExpanded,
			expectError:    false,
		},
		{
			name:           "Map StateExpanded with any subState should work",
			state:          indexertypes.StateExpanded,
			subState:       "any-substate",
			expectedResult: v2types.StateExpanded,
			expectError:    false,
		},
		{
			name:           "Map StateActive",
			state:          indexertypes.StateActive,
			subState:       indexertypes.SubStateTimelock, // subState doesn't matter for StateActive
			expectedResult: v2types.StateActive,
			expectError:    false,
		},
		{
			name:           "Map StatePending",
			state:          indexertypes.StatePending,
			subState:       indexertypes.SubStateTimelock, // subState doesn't matter for StatePending
			expectedResult: v2types.StatePending,
			expectError:    false,
		},
		{
			name:           "Map StateVerified",
			state:          indexertypes.StateVerified,
			subState:       indexertypes.SubStateTimelock, // subState doesn't matter for StateVerified
			expectedResult: v2types.StateVerified,
			expectError:    false,
		},
		{
			name:           "Map StateSlashed",
			state:          indexertypes.StateSlashed,
			subState:       indexertypes.SubStateTimelock, // subState doesn't matter for StateSlashed
			expectedResult: v2types.StateSlashed,
			expectError:    false,
		},
		{
			name:           "Map StateUnbonding with SubStateTimelock",
			state:          indexertypes.StateUnbonding,
			subState:       indexertypes.SubStateTimelock,
			expectedResult: v2types.StateTimelockUnbonding,
			expectError:    false,
		},
		{
			name:           "Map StateUnbonding with SubStateEarlyUnbonding",
			state:          indexertypes.StateUnbonding,
			subState:       indexertypes.SubStateEarlyUnbonding,
			expectedResult: v2types.StateEarlyUnbonding,
			expectError:    false,
		},
		{
			name:        "Map StateUnbonding with invalid subState should error",
			state:       indexertypes.StateUnbonding,
			subState:    "invalid-substate",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v2types.MapDelegationState(tt.state, tt.subState)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
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
			name: "Single param",
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: 5,
			expectError:    false,
		},
		{
			name: "Multiple params, select highest version",
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 3},
				{Version: 3, MaxFinalityProviders: 7}, // Highest version
				{Version: 2, MaxFinalityProviders: 5},
			},
			expectedResult: 7, // Version 3 is the highest
			expectError:    false,
		},
		{
			name: "Version order doesn't matter in input, highest selected",
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 5, MaxFinalityProviders: 10},
				{Version: 1, MaxFinalityProviders: 3},
				{Version: 3, MaxFinalityProviders: 7},
			},
			expectedResult: 10, // Version 5 is the highest
			expectError:    false,
		},
		{
			name:          "No params found",
			babylonParams: []*indexertypes.BbnStakingParams{},
			expectError:   true,
		},
		{
			name:        "Babylon params database error",
			paramsError: errors.New("database connection failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			indexerDB := mocks.NewIndexerDBClient(t)

			// Mock GetBbnStakingParams
			if tt.expectError && tt.paramsError != nil {
				indexerDB.On("GetBbnStakingParams", ctx).Return(nil, tt.paramsError).Once()
			} else {
				indexerDB.On("GetBbnStakingParams", ctx).Return(tt.babylonParams, nil).Once()
			}

			// Setup service
			cfg := &config.Config{}
			dbClients := &dbclients.DbClients{
				IndexerDBClient: indexerDB,
			}
			sharedService, err := service.New(cfg, nil, nil, nil, dbClients, &types.ChainInfo{
				ChainID: "babylon",
			})
			require.NoError(t, err)

			v2Service, err := New(sharedService, nil)
			require.NoError(t, err)

			// Execute
			result, err := v2Service.getLatestMaxFinalityProviders(ctx)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
