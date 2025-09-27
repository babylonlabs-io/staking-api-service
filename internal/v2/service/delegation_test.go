package v2service

import (
	"testing"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	v2types "github.com/babylonlabs-io/staking-api-service/internal/v2/types"
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
