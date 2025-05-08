//go:build e2e

package api

import (
	"net/http"
	"testing"
)

func TestV1_Delegation(t *testing.T) {
	cases := []testcase{
		{
			testName:         "missing parameter",
			endpoint:         "/v1/delegation",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"staking_tx_hash_hex is required"}`,
		},
		{
			testName:         "invalid staking_tx_hash_hex",
			endpoint:         "/v1/delegation?staking_tx_hash_hex=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid staking_tx_hash_hex"}`,
		},
		{
			testName:         "non existing staking_tx_hash_hex",
			endpoint:         "/v1/delegation?staking_tx_hash_hex=035929d276944fe1dd24aad43929d88be8e5c7b935757c59da08ba38df71238c",
			expectedHttpCode: http.StatusNotFound,
			expectedContents: `{"errorCode":"NOT_FOUND", "message":"staking delegation not found, please retry"}`,
		},
		{
			testName:         "ok",
			endpoint:         "/v1/delegation?staking_tx_hash_hex=19caaf9dcf7be81120a503b8e007189ecee53e5912c8fa542b187224ce45000a",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"finality_provider_pk_hex":"32009354f274871178dbb4ab7fa789f4a96fea8f0ff5de105b306c046e256769","is_eligible_for_transition":false,"is_overflow":false,"is_slashed":false,"staker_pk_hex":"21d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e06","staking_tx":{"output_index":0,"start_height":197535,"start_timestamp":"2024-05-28T12:19:08+04:00","timelock":64000,"tx_hex":"02000000000102ac8e1a0083ef08b08c32ab200d3aa74c0ba1130162f3f7c19bfc4636356662c90000000000fdffffff2a2ffd89feb0220ea44eab166eb9e2851ba09f24242267e5e74de5e71b03fbe50000000000fdffffff0220420200000000002251209ca9da0c4ef2aa56a794eee20a08e7d10f880204a5d5f5dd45151ed88c76e0290000000000000000496a47626274340021d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e0632009354f274871178dbb4ab7fa789f4a96fea8f0ff5de105b306c046e256769fa0001403bfd71bd791ba08739e727497c1b4dfb4f44165bb4347c47d16710b662fb66a4d5c4f6ef358e5971065c0896326fb2700c64f422d5255d84c2fee866db60c1810140d532ab0d0cafca7d758e2cc67c91acdd606400403ca2dcc4423edc6d533f21861f37d6b928330c9fda07dfe6f9c39dbefe784e7e27b90d213d458e330525d2c89e030300"},"staking_tx_hash_hex":"19caaf9dcf7be81120a503b8e007189ecee53e5912c8fa542b187224ce45000a","staking_value":148000,"state":"active"}}`,
		},
	}

	checkCases(t, cases)
}

func TestV1_FinalityProviders(t *testing.T) {
	t.Parallel()

	cases := []testcase{
		{
			testName:         "invalid fp_btc_pk",
			endpoint:         "/v1/finality-providers?fp_btc_pk=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid fp_btc_pk"}`,
		},
		{
			testName:         "invalid pagination key",
			endpoint:         "/v1/finality-providers?pagination_key=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid pagination key format"}`,
		},
		{
			testName:         "list all providers",
			endpoint:         "/v1/finality-providers",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":[{"description":{"moniker":"Babylon Foundation 2","identity":"","website":"","security_contact":"","details":""},"commission":"0.080000000000000000","btc_pk":"094f5861be4128861d69ea4b66a5f974943f100f55400bf26f5cce124b4c9af7","active_tvl":0,"total_tvl":0,"active_delegations":0,"total_delegations":0},{"description":{"moniker":"Babylon Foundation 1","identity":"","website":"","security_contact":"","details":""},"commission":"0.060000000000000000","btc_pk":"063deb187a4bf11c114cf825a4726e4c2c35fea5c4c44a20ff08a30a752ec7e0","active_tvl":0,"total_tvl":0,"active_delegations":0,"total_delegations":0},{"description":{"moniker":"Babylon Foundation 3","identity":"","website":"","security_contact":"","details":""},"commission":"0.090000000000000000","btc_pk":"0d2f9728abc45c0cdeefdd73f52a0e0102470e35fb689fc5bc681959a61b021f","active_tvl":0,"total_tvl":0,"active_delegations":0,"total_delegations":0},{"description":{"moniker":"Babylon Foundation 0","identity":"","website":"","security_contact":"","details":""},"commission":"0.050000000000000000","btc_pk":"03d5a0bb72d71993e435d6c5a70e2aa4db500a62cfaae33c56050deefee64ec0","active_tvl":0,"total_tvl":0,"active_delegations":0,"total_delegations":0}],"pagination":{"next_key":""}}`,
		},
		{
			testName:         "select specific finality provider",
			endpoint:         "/v1/finality-providers?fp_btc_pk=03d5a0bb72d71993e435d6c5a70e2aa4db500a62cfaae33c56050deefee64ec0",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":[{"active_delegations":0,"active_tvl":0,"btc_pk":"03d5a0bb72d71993e435d6c5a70e2aa4db500a62cfaae33c56050deefee64ec0","commission":"0.050000000000000000","description":{"details":"","identity":"","moniker":"Babylon Foundation 0","security_contact":"","website":""},"total_delegations":0,"total_tvl":0}]}`,
		},
		{
			testName:         "select non existing finality provider",
			endpoint:         "/v1/finality-providers?fp_btc_pk=88b32b005d5b7e29e6f82998aff023bff7b600c6a1a74ffac984b3aa0579b384",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data": null}`, // todo is it correct?
		},
	}

	checkCases(t, cases)
}

func TestV1_GlobalParams(t *testing.T) {
	contents := `{"data":{"versions":[{"activation_height":192840,"cap_height":0,"confirmation_depth":10,"covenant_pks":["0381b70c01535f5153a8039c21150c53f3e49a083555b57930103db8a7272ff336","02159f46467124f6bbba77060520571ddb07c7e95ff54d8b9958ec0b0d59d86c03","039705be04f3a3eb5c3d0dd61e648e06ea8170975744594fe702e8088bcceff375","02ce138027bfdfb4dd631e9cecf097082c8a505ab16de36f5e3eb816d105ba7575","03e15dba250612e79e22abf28a1828ba5e6bdfaaa6ed2d87462b046994c33fa46f"],"covenant_quorum":3,"max_staking_amount":1000000000,"max_staking_time":65000,"min_staking_amount":1000000,"min_staking_time":64000,"staking_cap":50000000000,"tag":"01020304","unbonding_fee":20000,"unbonding_time":1000,"version":0}]}}`
	assertResponse(t, "/v1/global-params", http.StatusOK, contents)
}

func TestV1_UnbondingEligibility(t *testing.T) {
	t.Parallel()

	cases := []testcase{
		{
			testName:         "missing parameter",
			endpoint:         "/v1/unbonding/eligibility",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"staking_tx_hash_hex is required"}`,
		},
		{
			testName:         "invalid staking_tx_hash_hex",
			endpoint:         "/v1/unbonding/eligibility?staking_tx_hash_hex=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid staking_tx_hash_hex"}`,
		},
		{
			testName:         "not found",
			endpoint:         "/v1/unbonding/eligibility?staking_tx_hash_hex=78503b1269fccaf05b00ea53df58eed0a9f614c88dc2170ea7dbcdd76e7cf202",
			expectedHttpCode: http.StatusForbidden,
			expectedContents: `{"errorCode":"NOT_FOUND", "message":"delegation not found"}`,
		},
		{
			testName:         "ok (not active)",
			endpoint:         "/v1/unbonding/eligibility?staking_tx_hash_hex=4eccd0df7dd7036bbd0771d5a47b7ab3b1ee396416c2d1c0cbfe3e7482564d14",
			expectedHttpCode: http.StatusForbidden,
			expectedContents: `{"errorCode":"FORBIDDEN", "message":"delegation state is not active"}`,
		},
		{
			testName:         "ok",
			endpoint:         "/v1/unbonding/eligibility?staking_tx_hash_hex=19caaf9dcf7be81120a503b8e007189ecee53e5912c8fa542b187224ce45000a",
			expectedHttpCode: http.StatusOK,
			expectedContents: `null`, // todo is it ok?
		},
	}
	checkCases(t, cases)
}

func TestV1_Stats(t *testing.T) {
	cases := []testcase{
		{
			testName:         "ok",
			endpoint:         "/v1/stats",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"active_delegations":417254,"active_tvl":67986511595,"pending_tvl":0,"total_delegations":478967,"total_stakers":343480,"total_tvl":72583507351,"unconfirmed_tvl":0}}`,
		},
	}

	checkCases(t, cases)
}

func TestV1_StakerDelegations(t *testing.T) {
	cases := []testcase{
		{
			testName:         "missing parameters",
			endpoint:         "/v1/staker/delegations",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"staker_btc_pk is required"}`,
		},
		{
			testName:         "invalid staker_btc_pk",
			endpoint:         "/v1/staker/delegations?staker_btc_pk=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid staker_btc_pk"}`,
		},
		{
			testName:         "non existing staker_btc_pk",
			endpoint:         "/v1/staker/delegations?staker_btc_pk=8b957fee1fadc87debe176fa8eb77b82c47b1224e26396d24eab0f892b2b1a45",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":[], "pagination":{"next_key":""}}`,
		},
		{
			testName:         "ok",
			endpoint:         "/v1/staker/delegations?staker_btc_pk=21d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e06",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":[{"staking_tx_hash_hex":"19caaf9dcf7be81120a503b8e007189ecee53e5912c8fa542b187224ce45000a","staker_pk_hex":"21d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e06","finality_provider_pk_hex":"32009354f274871178dbb4ab7fa789f4a96fea8f0ff5de105b306c046e256769","state":"active","staking_value":148000,"staking_tx":{"tx_hex":"02000000000102ac8e1a0083ef08b08c32ab200d3aa74c0ba1130162f3f7c19bfc4636356662c90000000000fdffffff2a2ffd89feb0220ea44eab166eb9e2851ba09f24242267e5e74de5e71b03fbe50000000000fdffffff0220420200000000002251209ca9da0c4ef2aa56a794eee20a08e7d10f880204a5d5f5dd45151ed88c76e0290000000000000000496a47626274340021d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e0632009354f274871178dbb4ab7fa789f4a96fea8f0ff5de105b306c046e256769fa0001403bfd71bd791ba08739e727497c1b4dfb4f44165bb4347c47d16710b662fb66a4d5c4f6ef358e5971065c0896326fb2700c64f422d5255d84c2fee866db60c1810140d532ab0d0cafca7d758e2cc67c91acdd606400403ca2dcc4423edc6d533f21861f37d6b928330c9fda07dfe6f9c39dbefe784e7e27b90d213d458e330525d2c89e030300","output_index":0,"start_timestamp":"2024-05-28T12:19:08+04:00","start_height":197535,"timelock":64000},"is_overflow":false,"is_eligible_for_transition":false,"is_slashed":false}],"pagination":{"next_key":""}}`,
		},
	}

	checkCases(t, cases)
}
