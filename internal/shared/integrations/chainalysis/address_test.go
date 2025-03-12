package chainalysis_test

import (
	"context"
	"fmt"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainalysis"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	// temporary while we won't transition to non global metrics
	metrics.Init(9999)

	ctx := context.Background()

	const apiKey = "valid_api_key"
	const (
		noRiskAddress = "bc1qcczn6c9535nnklry4qymadthryvu3sfg2ndsve"
		riskAddress   = "12NpCkhddSNiDkD9rRYUCHsTT9ReMNiJjG"
	)

	// only important fields were taken from real response to the api with specified address
	srv := setupTestServer(t, apiKey, map[string]string{riskAddress: `{"address":"12NpCkhddSNiDkD9rRYUCHsTT9ReMNiJjG","risk":"Severe","cluster":{},"riskReason":"Identified as Sanctioned Entity","addressType":"PRIVATE_WALLET","addressIdentifications":[],"exposures":[],"triggers":[],"status":"COMPLETE"}`})

	t.Run("Invalid api key", func(t *testing.T) {
		client := chainalysis.NewClient("invalid_api_key", srv.URL)
		resp, err := client.AssessAddress(ctx, noRiskAddress)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	t.Run("Low risk address", func(t *testing.T) {
		client := chainalysis.NewClient(apiKey, srv.URL)
		resp, err := client.AssessAddress(ctx, noRiskAddress)
		require.NoError(t, err)

		assert.Equal(t, "Low", resp.Risk)
		assert.Nil(t, resp.RiskReason)
	})
	t.Run("Severe risk address", func(t *testing.T) {
		client := chainalysis.NewClient(apiKey, srv.URL)
		resp, err := client.AssessAddress(ctx, riskAddress)
		require.NoError(t, err)

		assert.Equal(t, "Severe", resp.Risk)
		assert.Equal(t, "Identified as Sanctioned Entity", *resp.RiskReason)
	})
	t.Run("Internal server error", func(t *testing.T) {
		// mind nil as last parameter - it will trigger 5xx
		srv := setupTestServer(t, apiKey, nil)
		client := chainalysis.NewClient(apiKey, srv.URL)
		resp, err := client.AssessAddress(ctx, noRiskAddress)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// setupTestServer starts http server for tests that tries to mimic chainalysis API
// expectedToken is token that should be used in the request
// responses is map of address to response that should be returned for that address, if it's nil server returns 500 error
func setupTestServer(t *testing.T, expectedToken string, responses map[string]string) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Token")
		if token != expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// for now we use only one endpoint, having this simple check is sufficient
		address, correctPath := strings.CutPrefix(r.URL.Path, "/api/risk/v2/entities/")
		if !correctPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if responses == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var err error
		if response, ok := responses[address]; ok {
			_, err = w.Write([]byte(response))
		} else {
			// even if requested address doesn't exist their API assumes that address has low risk
			response = fmt.Sprintf(`{"address":"%s","risk":"Low","cluster":null,"riskReason":null,"addressType":"PRIVATE_WALLET","addressIdentifications":[],"exposures":[],"triggers":[],"status":"COMPLETE"}`, address)
			_, err = w.Write([]byte(response))
		}
		require.NoError(t, err)
	}))
	t.Cleanup(srv.Close)

	return srv
}
