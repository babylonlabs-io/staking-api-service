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
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2types "github.com/babylonlabs-io/staking-api-service/internal/v2/types"
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
	testHash8, err := testutil.RandomAlphaNum(10)
	require.NoError(t, err)
	testHash9, err := testutil.RandomAlphaNum(10)
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
		{
			name: "Expanded delegation should not expand further",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateExpanded,
				StakingTxHashHex:          testHash8,
				FinalityProviderBtcPksHex: []string{"fp1", "fp2"},
			},
			allowList: map[string]bool{
				testHash8: true,
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false, // Already expanded, should not expand further
		},
		{
			name: "Expanded delegation not in allow-list should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateExpanded,
				StakingTxHashHex:          testHash9,
				FinalityProviderBtcPksHex: []string{"fp1", "fp2"},
			},
			allowList: map[string]bool{
				"other-hash": true,
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false, // Not in allow-list, should not expand
		},
		{
			name: "Active expanded delegation should check allowlist using original hash - should expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "expanded_hash_123", // New expanded delegation hash
				PreviousStakingTxHashHex:  testHash1,           // Original hash that's in allowlist
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				testHash1: true, // Original hash is in allowlist, not the expanded hash
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true, // Should expand because original hash is in allowlist
		},
		{
			name: "Active expanded delegation where original hash is not in allowlist - should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "expanded_hash_456", // New expanded delegation hash
				PreviousStakingTxHashHex:  "original_hash_not_in_allowlist",
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"some_other_hash": true, // Original hash is NOT in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false, // Should not expand because original hash is not in allowlist
		},
		{
			name: "Active expanded delegation with original hash NOT in allowlist but current expanded hash IS in allowlist - should expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "expanded_hash_in_allowlist",     // Current expanded hash IS in allowlist
				PreviousStakingTxHashHex:  "original_hash_not_in_allowlist", // Original hash NOT in allowlist
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"expanded_hash_in_allowlist": true, // Current expanded hash is in allowlist
				"some_other_hash":            true, // Original hash is NOT in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true, // Should expand because current expanded hash is in allowlist
		},
		{
			name: "Active expanded delegation with BOTH original and current hash in allowlist - should expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "expanded_hash_also_in_allowlist", // Current expanded hash IS in allowlist
				PreviousStakingTxHashHex:  testHash2,                         // Original hash IS in allowlist
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				testHash2:                         true, // Original hash is in allowlist
				"expanded_hash_also_in_allowlist": true, // Current expanded hash is also in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true, // Should expand because both hashes are in allowlist
		},
		{
			name: "Active expanded delegation with NEITHER original nor current hash in allowlist - should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "expanded_hash_not_in_allowlist",      // Current expanded hash NOT in allowlist
				PreviousStakingTxHashHex:  "original_hash_also_not_in_allowlist", // Original hash NOT in allowlist
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"completely_different_hash": true, // Neither hash is in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false, // Should not expand because neither hash is in allowlist
		},
		{
			name: "Chained expansion: delegation1->delegation2->delegation3, only delegation1 in allowlist - should expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "delegation3_hash", // Current (delegation3)
				PreviousStakingTxHashHex:  "delegation2_hash", // Points to delegation2 (immediate previous)
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				testHash1: true, // Only delegation1 (original) is in allowlist
				// delegation2_hash and delegation3_hash are NOT in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true, // Should expand because delegation1 (root) is in allowlist
		},
		{
			name: "Deep chain: delegation at depth 5 with root in allowlist - should expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "delegation5_hash", // Current (delegation5)
				PreviousStakingTxHashHex:  "delegation4_hash", // Points to delegation4
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				testHash2: true, // Only delegation1 (original) is in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true, // Should expand because delegation1 (root) is in allowlist
		},
		{
			name: "Cycle detection: delegation with circular reference - should not expand",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "delegationA_hash", // Current (delegationA)
				PreviousStakingTxHashHex:  "delegationB_hash", // Points to delegationB
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"some_hash": true, // None of the cycle hashes are in allowlist
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false, // Should not expand due to cycle detection
		},

		{
			name: "Chain expansion: root allowlisted",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "del3",
				PreviousStakingTxHashHex:  "del2",
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"del1": true, // Only root is allowlisted
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true,
		},

		{
			name: "Chain expansion: middle allowlisted",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "del3",
				PreviousStakingTxHashHex:  "del2",
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"del2": true, // Middle delegation allowlisted
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true,
		},

		{
			name: "Chain expansion: leaf allowlisted",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "del3",
				PreviousStakingTxHashHex:  "del2",
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"del3": true, // Current (leaf) delegation allowlisted
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: true,
		},

		{
			name: "Chain expansion: none allowlisted",
			delegation: indexerdbmodel.IndexerDelegationDetails{
				State:                     indexertypes.StateActive,
				StakingTxHashHex:          "del3",
				PreviousStakingTxHashHex:  "del2",
				FinalityProviderBtcPksHex: []string{"fp1"},
			},
			allowList: map[string]bool{
				"other": true, // None in chain allowlisted
			},
			babylonParams: []*indexertypes.BbnStakingParams{
				{Version: 1, MaxFinalityProviders: 5},
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexerDB := mocks.NewIndexerDBClient(t)
			v1DB := mocks.NewV1DBClient(t)

			// Only mock GetBbnStakingParams call if delegation is active (will reach the params check)
			if tt.delegation.State == indexertypes.StateActive {
				if tt.expectedErrorInGetParams {
					indexerDB.On("GetBbnStakingParams", ctx).Return(nil, tt.babylonParamsError).Once()
				} else {
					indexerDB.On("GetBbnStakingParams", ctx).Return(tt.babylonParams, nil).Once()
				}

				// Mock chain traversal calls for specific test cases
				switch tt.name {
				case "Chained expansion: delegation1->delegation2->delegation3, only delegation1 in allowlist - should expand":
					// Mock GetDelegation call for delegation2_hash -> return delegation2
					delegation2 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "delegation2_hash",
						PreviousStakingTxHashHex:  testHash1, // Points to delegation1 (which is in allowlist)
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "delegation2_hash").Return(delegation2, nil).Once()

					// Mock GetDelegation call for testHash1 (delegation1) -> return delegation1 (root)
					delegation1 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          testHash1,
						PreviousStakingTxHashHex:  "", // Root delegation, no previous
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, testHash1).Return(delegation1, nil).Once()

				case "Deep chain: delegation at depth 5 with root in allowlist - should expand":
					// Mock the chain: delegation5 -> delegation4 -> delegation3 -> delegation2 -> delegation1 (testHash2)
					delegation4 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "delegation4_hash",
						PreviousStakingTxHashHex:  "delegation3_hash",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "delegation4_hash").Return(delegation4, nil).Once()

					delegation3 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "delegation3_hash",
						PreviousStakingTxHashHex:  "delegation2_hash_deep",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "delegation3_hash").Return(delegation3, nil).Once()

					delegation2Deep := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "delegation2_hash_deep",
						PreviousStakingTxHashHex:  testHash2, // Points to delegation1 (root, in allowlist)
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "delegation2_hash_deep").Return(delegation2Deep, nil).Once()

					delegation1Deep := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          testHash2,
						PreviousStakingTxHashHex:  "", // Root delegation, no previous
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, testHash2).Return(delegation1Deep, nil).Once()

				case "Cycle detection: delegation with circular reference - should not expand":
					// Mock the cycle: delegationA -> delegationB -> delegationA (cycle)
					delegationB := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "delegationB_hash",
						PreviousStakingTxHashHex:  "delegationA_hash", // This creates the cycle
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "delegationB_hash").Return(delegationB, nil).Once()

				case "Active expanded delegation should check allowlist using original hash - should expand":
					// This test has PreviousStakingTxHashHex: testHash1 which is in allowlist
					// No need to traverse further since testHash1 is directly in allowlist
					break

				case "Active expanded delegation where original hash is not in allowlist - should not expand":
					// Mock the chain: expanded_hash_456 -> original_hash_not_in_allowlist (root, not allowlisted)
					originalDelegation := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "original_hash_not_in_allowlist",
						PreviousStakingTxHashHex:  "", // Root delegation
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "original_hash_not_in_allowlist").Return(originalDelegation, nil).Once()

				case "Active expanded delegation with NEITHER original nor current hash in allowlist - should not expand":
					// Mock the chain: expanded_hash_not_in_allowlist -> original_hash_also_not_in_allowlist (both not allowlisted)
					originalNotAllowlisted := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "original_hash_also_not_in_allowlist",
						PreviousStakingTxHashHex:  "", // Root delegation
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "original_hash_also_not_in_allowlist").Return(originalNotAllowlisted, nil).Once()

				case "Chain expansion: root allowlisted":
					// Mock: del3 -> del2 -> del1 (root allowlisted)
					del2 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "del2",
						PreviousStakingTxHashHex:  "del1",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "del2").Return(del2, nil).Once()

					del1 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "del1",
						PreviousStakingTxHashHex:  "",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "del1").Return(del1, nil).Once()

				case "Chain expansion: middle allowlisted":
					// Mock: del3 -> del2 (allowlisted)
					del2 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "del2",
						PreviousStakingTxHashHex:  "del1",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "del2").Return(del2, nil).Once()

				case "Chain expansion: leaf allowlisted":
					// Current delegation (del3) is allowlisted, no traversal needed
					break

				case "Chain expansion: none allowlisted":
					// Mock: del3 -> del2 -> del1 (none allowlisted)
					del2 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "del2",
						PreviousStakingTxHashHex:  "del1",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "del2").Return(del2, nil).Once()

					del1 := &indexerdbmodel.IndexerDelegationDetails{
						StakingTxHashHex:          "del1",
						PreviousStakingTxHashHex:  "",
						State:                     indexertypes.StateExpanded,
						FinalityProviderBtcPksHex: []string{"fp1"},
					}
					indexerDB.On("GetDelegation", ctx, "del1").Return(del1, nil).Once()
				}
			}

			cfg := &config.Config{}
			dbClients := &dbclients.DbClients{
				IndexerDBClient: indexerDB,
				V1DBClient:      v1DB,
			}

			sharedService, err := service.New(cfg, nil, nil, nil, dbClients, &types.ChainInfo{
				ChainID: "babylon",
			})
			require.NoError(t, err)

			v2Service, err := New(sharedService, nil, tt.allowList)
			require.NoError(t, err)

			result := v2Service.evaluateCanExpand(ctx, tt.delegation)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

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

			v2Service, err := New(sharedService, nil, nil)
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
