package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFinalityProviderState(t *testing.T) {
	tests := []struct {
		name           string
		dbState        string
		expectedResult FinalityProviderQueryingState
	}{
		{
			name:           "Map FINALITY_PROVIDER_STATUS_ACTIVE to active",
			dbState:        "FINALITY_PROVIDER_STATUS_ACTIVE",
			expectedResult: FinalityProviderStateActive,
		},
		{
			name:           "Map FINALITY_PROVIDER_STATUS_INACTIVE to standby",
			dbState:        "FINALITY_PROVIDER_STATUS_INACTIVE",
			expectedResult: FinalityProviderStateStandby,
		},
		{
			name:           "Map FINALITY_PROVIDER_STATUS_JAILED to standby",
			dbState:        "FINALITY_PROVIDER_STATUS_JAILED",
			expectedResult: FinalityProviderStateStandby,
		},
		{
			name:           "Map FINALITY_PROVIDER_STATUS_SLASHED to standby",
			dbState:        "FINALITY_PROVIDER_STATUS_SLASHED",
			expectedResult: FinalityProviderStateStandby,
		},
		{
			name:           "Map empty string to standby",
			dbState:        "",
			expectedResult: FinalityProviderStateStandby,
		},
		{
			name:           "Map unexpected value to standby",
			dbState:        "UNKNOWN_STATUS",
			expectedResult: FinalityProviderStateStandby,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFinalityProviderState(tt.dbState)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
