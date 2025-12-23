//go:build e2e

package api

import (
	"net/http"
	"testing"
)

func TestV2_FinalityProviders(t *testing.T) {
	// Invalid pagination key format
	assertResponse(t, "/v2/finality-providers?pagination_key=invalid", http.StatusBadRequest, `{"errorCode":"BAD_REQUEST", "message":"invalid pagination key format"}`)

	// First page: returns 2 providers (MaxPaginationLimit=2) with pagination token
	firstPageContents := `{"data":[{"btc_pk":"bef341a7adb10213a7ec7825afeb7d57fbfa7b5f7bdf201204fb0ef62fb9cfa6","state":"FINALITY_PROVIDER_STATUS_ACTIVE","description":{"moniker":"verse2","identity":"","website":"https://verse2.io","security_contact":"ted@verse2.io","details":""},"commission":"0.050000000000000000","active_tvl":0,"active_delegations":0,"type":""},{"btc_pk":"d23c2c25e1fcf8fd1c21b9a402c19e2e309e531e45e92fb1e9805b6056b0cc76","state":"FINALITY_PROVIDER_STATUS_ACTIVE","description":{"moniker":"Babylon Foundation 0","identity":"","website":"https://babylonlabs.io","security_contact":"","details":""},"commission":"0.100000000000000000","active_tvl":0,"active_delegations":0,"type":""}],"pagination":{"next_key":"eyJidGNfcGsiOiJkMjNjMmMyNWUxZmNmOGZkMWMyMWI5YTQwMmMxOWUyZTMwOWU1MzFlNDVlOTJmYjFlOTgwNWI2MDU2YjBjYzc2In0="}}`
	assertResponse(t, "/v2/finality-providers", http.StatusOK, firstPageContents)

	// Second page: returns remaining 1 provider with empty pagination token
	secondPageContents := `{"data":[{"btc_pk":"e4889630fa8695dae630c41cd9b85ef165ccc2dc5e5935d5a24393a9defee9ef","state":"FINALITY_PROVIDER_STATUS_ACTIVE","description":{"moniker":"Babylon Foundation 1","identity":"","website":"https://babylonlabs.io","security_contact":"","details":""},"commission":"0.070000000000000000","active_tvl":0,"active_delegations":0,"type":""}],"pagination":{"next_key":""}}`
	assertResponse(t, "/v2/finality-providers?pagination_key=eyJidGNfcGsiOiJkMjNjMmMyNWUxZmNmOGZkMWMyMWI5YTQwMmMxOWUyZTMwOWU1MzFlNDVlOTJmYjFlOTgwNWI2MDU2YjBjYzc2In0=", http.StatusOK, secondPageContents)
}

func TestV2_NetworkInfo(t *testing.T) {
	cases := []testcase{
		{
			testName:         "ok",
			endpoint:         "/v2/network-info",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"params":{"bbn":[{"version":0,"covenant_pks":["49766ccd9e3cd94343e2040474a77fb37cdfd30530d05f9f1e96ae1e2102c86e","76d1ae01f8fb6bf30108731c884cddcf57ef6eef2d9d9559e130894e0e40c62c","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":5000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":2000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":197535},{"version":1,"covenant_pks":["09585ab55a971a231c945790a0a81df754e5a07263a5c20829931cc24683bbb7","76d1ae01f8fb6bf30108731c884cddcf57ef6eef2d9d9559e130894e0e40c62c","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":5000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":10000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":198665},{"version":2,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":5000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":10000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":200665},{"version":3,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":50000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":5000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":215968},{"version":4,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":50000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":5000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":220637},{"version":5,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":35000000000,"min_staking_time_blocks":10000,"max_staking_time_blocks":64000,"slashing_pk_script":"00145be12624d08a2b424095d7c07221c33450d14bf1","min_slashing_tx_fee_sat":5000,"slashing_rate":"0.050000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":2000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1095000,"allow_list_expiration_height":26124,"btc_activation_height":227174},{"version":6,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":35000000000,"min_staking_time_blocks":10000,"max_staking_time_blocks":64000,"slashing_pk_script":"00145be12624d08a2b424095d7c07221c33450d14bf1","min_slashing_tx_fee_sat":6000,"slashing_rate":"0.050000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":2000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1095000,"allow_list_expiration_height":26124,"btc_activation_height":235952}],"btc":[{"version":0,"btc_confirmation_depth":10,"checkpoint_finalization_timeout":100}]}}}`,
		},
	}

	checkCases(t, cases)
}

func TestV2_Prices(t *testing.T) {
	contents := `{"errorCode":"INTERNAL_SERVICE_ERROR","message":"Internal service error"}`
	assertResponse(t, "/v2/prices", http.StatusInternalServerError, contents)
}

func TestV2_Stats(t *testing.T) {
	contents := `{"data":{"active_tvl":3000,"active_delegations":3000,"active_finality_providers":3,"total_finality_providers":3,"total_active_tvl":67986514595,"total_active_delegations":420254,"btc_staking_apr":0,"max_staking_apr":0}}`
	assertResponse(t, "/v2/stats", http.StatusOK, contents)
}

func TestV2_AddressScreening(t *testing.T) {
	cases := []testcase{
		{
			testName:         "missing parameter",
			endpoint:         "/address/screening",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST","message":"btc_address is required"}`,
		},
		{
			testName:         "invalid btc_address",
			endpoint:         "/address/screening?btc_address=invalid",
			expectedHttpCode: http.StatusInternalServerError,
			expectedContents: `{"errorCode":"INTERNAL_SERVICE_ERROR","message":"Internal service error"}`,
		},
	}

	checkCases(t, cases)
}

func TestV2_Delegation(t *testing.T) {
	cases := []testcase{
		{
			testName:         "missing parameter",
			endpoint:         "/v2/delegation",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"staking_tx_hash_hex is required"}`,
		},
		{
			testName:         "invalid staking_tx_hash_hex",
			endpoint:         "/v2/delegation?staking_tx_hash_hex=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid staking_tx_hash_hex"}`,
		},
		{
			testName:         "non existing staking_tx_hash_hex",
			endpoint:         "/v2/delegation?staking_tx_hash_hex=61325403c5a553f7e5a061d314904c02a9ec4202a2616b531335998b4506d43b",
			expectedHttpCode: http.StatusNotFound,
			expectedContents: `{"errorCode":"NOT_FOUND", "message":"staking delegation not found, please retry"}`,
		},
	}

	checkCases(t, cases)
}

func TestV2_Delegations(t *testing.T) {
	cases := []testcase{
		{
			testName:         "missing parameters",
			endpoint:         "/v2/delegations",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"staker_pk_hex or babylon_address is required"}`,
		},
		{
			testName:         "invalid staker_pk_hex",
			endpoint:         "/v2/delegations?staker_pk_hex=invalid",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"invalid staker_pk_hex"}`,
		},
		{
			testName:         "non existing staker_pk_hex",
			endpoint:         "/v2/delegations?staker_pk_hex=5f61c497c58a8961bc3a0d493971773635f928261596944c2f7e9f565114c6f3",
			expectedHttpCode: http.StatusOK, // todo is it ok?
			expectedContents: `{"data":[],"pagination":{"next_key":""}}`,
		},
	}

	checkCases(t, cases)
}

func TestV2_StakerStats(t *testing.T) {
	cases := []testcase{
		{
			testName:         "missing parameters",
			endpoint:         "/v2/staker/stats",
			expectedHttpCode: http.StatusBadRequest,
			expectedContents: `{"errorCode":"BAD_REQUEST","message":"staker_pk_hex is required"}`,
		},
		{
			testName:         "ok",
			endpoint:         "/v2/staker/stats?staker_pk_hex=1e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"active_tvl":50000,"active_delegations":1,"unbonding_tvl":0,"unbonding_delegations":0,"withdrawable_tvl":0,"withdrawable_delegations":0,"slashed_tvl":0,"slashed_delegations":0}}`,
		},
		{
			testName:         "existing stats and empty babylon_address",
			endpoint:         "/v2/staker/stats?staker_pk_hex=1e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1&babylon_address=",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"active_tvl":50000,"active_delegations":1,"unbonding_tvl":0,"unbonding_delegations":0,"withdrawable_tvl":0,"withdrawable_delegations":0,"slashed_tvl":0,"slashed_delegations":0}}`,
		},
	}

	checkCases(t, cases)
}
