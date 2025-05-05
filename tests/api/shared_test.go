//go:build e2e

package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestV1_StakerPubkeyLookup(t *testing.T) {
	t.Parallel()

	cases := []testcase{
		{
			testName:         "no address provided",
			endpoint:         "/v1/staker/pubkey-lookup",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"address is required"}`,
		},
		{
			testName:         "invalid address",
			endpoint:         "/v1/staker/pubkey-lookup?address=invalid_addr",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"can not decode btc address: decoded address is of unknown format"}`,
		},
		{
			testName:         "non existing public key",
			endpoint:         "/v1/staker/pubkey-lookup?address=bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{}}`,
		},
		{
			testName: "exceeding max addresses",
			// generating long query
			endpoint:         "/v1/staker/pubkey-lookup?" + strings.Repeat("address=addr1&", 100),
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"Maximum 10 address allowed"}`,
		},
	}
	checkCases(t, cases)

	t.Run("ok", func(t *testing.T) {
		addresses := []string{
			"bc1pwwrdfkea9qtp5fg9m630c403vq950s0s36h7pz6435mgafl7auls2cpnvh", // taproot
			"bc1qyepnd0hvyxakv3xmh48q5hv4plfsqp8vq9a7uz",                     // native segwit even, but corresponds to the same doc as above
			"bc1qfnqcsx5pk9ct4v4v4e58z8wekkw2lm0fqx9cf5",                     // native segwit even
			"bc1qwrcdqzme084nnkhgjtgerj3dcfyntgz6q80tsq",                     // native segwit odd
		}

		var query string
		for _, addr := range addresses {
			query += fmt.Sprintf("address=%s&", addr)
		}
		endpoint := "/v1/staker/pubkey-lookup?" + query

		const expected = `
		{"data":{
			"bc1pwwrdfkea9qtp5fg9m630c403vq950s0s36h7pz6435mgafl7auls2cpnvh":"3faa3aa676b2addc8e4750b65d02f54386e0dbc87d83cdbf3bd02053b0ed0bcf",
			"bc1qfnqcsx5pk9ct4v4v4e58z8wekkw2lm0fqx9cf5":"1e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1",
			"bc1qwrcdqzme084nnkhgjtgerj3dcfyntgz6q80tsq":"c00f83fb8dbed188175c67937c1d62f20eaf04a995b978e72f025c9c980f7a5d",
			"bc1qyepnd0hvyxakv3xmh48q5hv4plfsqp8vq9a7uz":"3faa3aa676b2addc8e4750b65d02f54386e0dbc87d83cdbf3bd02053b0ed0bcf"
		}}
		`
		assertResponse(t, endpoint, http.StatusOK, expected)
	})
}

func TestV1_StakerDelegationCheck(t *testing.T) {
	cases := []testcase{
		{
			testName:         "no params provided",
			endpoint:         "/v1/staker/delegation/check",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"address is required"}`,
		},
		{
			testName:         "invalid address",
			endpoint:         "/v1/staker/delegation/check?address=invalid_addr",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"can not decode btc address: decoded address is of unknown format"}`,
		},
		{
			testName:         "invalid timeframe",
			endpoint:         "/v1/staker/delegation/check?address=bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq&timeframe=tomorrow",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid timeframe value"}`,
		},
		{
			testName:         "non existing address with valid timeframe",
			endpoint:         "/v1/staker/delegation/check?address=bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq&timeframe=today",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"code":0, "data":false}`,
		},
		{
			testName:         "valid but not active delegation",
			endpoint:         "/v1/staker/delegation/check?address=tb1pxn93chqf33caw2dxs786leqqxwc36r603auj926347tc8n3rrjdssjcf6k",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"code":0, "data":false}`,
		},
		{
			// staking_btc_timestamp happened long time ago which mean it won't be covered by timeframe=today
			testName:         "valid delegation not covered by chosen timeframe",
			endpoint:         "/v1/staker/delegation/check?address=tb1pwa22ug5nsz76euemjqg7n9m9thasgtsp5h4q0n5cktn3dczm0hmq9fleyd&timeframe=today",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"code":0, "data":false}`,
		},
		{
			testName:         "ok",
			endpoint:         "/v1/staker/delegation/check?address=bc1p4uscanqkv7r3fc9kf3e9gs3jwe8zvhztxnfx8cmrgmm5p0txx8zsuygt0d",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"code":0, "data":true}`,
		},
	}
	checkCases(t, cases)
}
