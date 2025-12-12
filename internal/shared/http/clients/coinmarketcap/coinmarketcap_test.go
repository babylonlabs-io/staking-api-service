//go:build manual

package coinmarketcap

import (
	"os"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	key := os.Getenv("COINMARKETCAP_KEY")
	require.NotEmpty(t, key)

	metrics.Init(0)

	cl := NewClient(key, 5*time.Second)

	t.Run("ok", func(t *testing.T) {
		quotes, err := cl.LatestQuote(t.Context(), BtcID)
		require.NoError(t, err)
		spew.Dump(quotes)
	})
	t.Run("non existing id", func(t *testing.T) {
		const nonExistingID = 100000000
		quotes, err := cl.LatestQuote(t.Context(), nonExistingID)
		require.Error(t, err)
		assert.Nil(t, quotes)
	})
}
