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
	assert.True(t, isValidURL(logoURL))
}

func isValidURL(urlStr string) bool {
	parsedURL, err := url.ParseRequestURI(urlStr)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}
