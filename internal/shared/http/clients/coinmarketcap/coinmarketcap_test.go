//go:build manual

package coinmarketcap

import (
	"os"
	"testing"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestMe(t *testing.T) {
	key := os.Getenv("COINMARKETCAP_KEY")
	require.NotEmpty(t, key)

	metrics.Init(0)

	cl := NewClient(key, 5000)
	quotes, err := cl.LatestQuotes(t.Context(), 100000000)
	require.NoError(t, err)
	spew.Dump(quotes)
}
