package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHomePath = "/home/testuser"

func TestGetDefaultConfigFile(t *testing.T) {
	t.Run("returns correct file path", func(t *testing.T) {
		homePath := testHomePath
		filename := "test.json"

		expected := testHomePath + "/test.json"
		result := getDefaultConfigFile(homePath, filename)

		assert.Equal(t, expected, result)
	})

	t.Run("handles empty filename", func(t *testing.T) {
		homePath := testHomePath
		filename := ""

		expected := testHomePath
		result := getDefaultConfigFile(homePath, filename)

		assert.Equal(t, expected, result)
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("returns set value", func(t *testing.T) {
		// Set the global variable
		cfgPath = "/path/to/config.yml"

		result := GetConfigPath()
		assert.Equal(t, "/path/to/config.yml", result)

		// Reset for other tests
		cfgPath = ""
	})
}

func TestGetGlobalParamsPath(t *testing.T) {
	t.Run("returns set value", func(t *testing.T) {
		// Set the global variable
		globalParamsPath = "/path/to/params.json"

		result := GetGlobalParamsPath()
		assert.Equal(t, "/path/to/params.json", result)

		// Reset for other tests
		globalParamsPath = ""
	})
}

func TestGetFinalityProvidersPath(t *testing.T) {
	t.Run("returns set value", func(t *testing.T) {
		// Set the global variable
		finalityProvidersPath = "/path/to/providers.json"

		result := GetFinalityProvidersPath()
		assert.Equal(t, "/path/to/providers.json", result)

		// Reset for other tests
		finalityProvidersPath = ""
	})
}

func TestGetReplayFlag(t *testing.T) {
	t.Run("returns false by default", func(t *testing.T) {
		// Reset the global variable
		replayFlag = false

		result := GetReplayFlag()
		assert.False(t, result)
	})

	t.Run("returns true when set", func(t *testing.T) {
		// Set the global variable
		replayFlag = true

		result := GetReplayFlag()
		assert.True(t, result)

		// Reset for other tests
		replayFlag = false
	})
}

func TestGetBackfillPubkeyAddressFlag(t *testing.T) {
	t.Run("returns false by default", func(t *testing.T) {
		// Reset the global variable
		backfillPubkeyAddressFlag = false

		result := GetBackfillPubkeyAddressFlag()
		assert.False(t, result)
	})

	t.Run("returns true when set", func(t *testing.T) {
		// Set the global variable
		backfillPubkeyAddressFlag = true

		result := GetBackfillPubkeyAddressFlag()
		assert.True(t, result)

		// Reset for other tests
		backfillPubkeyAddressFlag = false
	})
}

func TestSetupFunction(t *testing.T) {
	t.Run("Setup configures all expected flags", func(t *testing.T) {
		// Reset global variables
		resetGlobalVariables()

		// Store original rootCmd and restore it after test
		originalRootCmd := rootCmd
		defer func() { rootCmd = originalRootCmd }()

		// Create a test command that we can inspect
		testCmd := &cobra.Command{
			Use: "test",
		}
		rootCmd = testCmd

		// Call Setup to configure flags
		_ = Setup()

		// Verify all expected flags exist
		expectedFlags := []string{"config", "params", "finality-providers", "replay", "backfill-pubkey-address"}
		for _, flagName := range expectedFlags {
			flag := testCmd.PersistentFlags().Lookup(flagName)
			require.NotNil(t, flag, "flag %s should exist", flagName)
		}
	})
}

// Helper function to reset global variables for testing
func resetGlobalVariables() {
	cfgPath = ""
	globalParamsPath = ""
	finalityProvidersPath = ""
	replayFlag = false
	backfillPubkeyAddressFlag = false
}
