package coinmarketcap

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMe(t *testing.T) {
	metrics.Init(0)

	cl := NewClient("d008a647-236a-40c7-bab8-09354ba05391", 5000)
	quotes, err := cl.LatestQuotes(t.Context(), "BTC")
	require.NoError(t, err)
	spew.Dump(quotes)
}
