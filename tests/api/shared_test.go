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
			"tb1pjpa390lrt7cl6ge4gs7kx7auaxqkdaj9jqzvqh5drwawa8e5j9ks2ldcpn", // taproot
			"tb1q663tkc4deys62yumazf9ht8a6cn44j8xpa75p6",                     // native segwit even, but corresponds to the same doc as above
			"tb1qp3hlkjuf80dxxpflvvwhl7h0zyxzne5zgtyc4v",                     // native segwit even
			"tb1qkgh4sawexmxzuwdrffxc9jjrnrpyj5x4nvy637",                     // native segwit odd
		}

		var query string
		for _, addr := range addresses {
			query += fmt.Sprintf("address=%s&", addr)
		}
		endpoint := "/v1/staker/pubkey-lookup?" + query

		const expected = `
		{"data":{
			"tb1pjpa390lrt7cl6ge4gs7kx7auaxqkdaj9jqzvqh5drwawa8e5j9ks2ldcpn":"3faa3aa676b2addc8e4750b65d02f54386e0dbc87d83cdbf3bd02053b0ed0bcf",
			"tb1q663tkc4deys62yumazf9ht8a6cn44j8xpa75p6": "3faa3aa676b2addc8e4750b65d02f54386e0dbc87d83cdbf3bd02053b0ed0bcf",
			"tb1qkgh4sawexmxzuwdrffxc9jjrnrpyj5x4nvy637": "c00f83fb8dbed188175c67937c1d62f20eaf04a995b978e72f025c9c980f7a5d",
			"tb1qp3hlkjuf80dxxpflvvwhl7h0zyxzne5zgtyc4v": "1e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1"
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
			endpoint:         "/v1/staker/delegation/check?address=tb1pwa22ug5nsz76euemjqg7n9m9thasgtsp5h4q0n5cktn3dczm0hmq9fleyd",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"code":0, "data":true}`,
		},
	}
	checkCases(t, cases)
}
