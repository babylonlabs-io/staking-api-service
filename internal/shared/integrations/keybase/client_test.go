//go:build manual

package keybase

import (
	"testing"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeybaseClient(t *testing.T) {
	metrics.Init(0)

	client := NewClient()
	logoURL, err := client.GetLogoURL(t.Context(), "83D300CB42D06962")
	require.NoError(t, err)
	assert.Equal(t, "https://s3.amazonaws.com/keybase_processed_uploads/1c7c29dec05c920a99b42e114e732705_360_360.jpg", logoURL)
}
