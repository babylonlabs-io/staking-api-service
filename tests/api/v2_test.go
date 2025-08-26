//go:build e2e

package api

import (
	"net/http"
	"testing"
)

func TestV2_FinalityProviders(t *testing.T) {
	contents := `{"data":[{"btc_pk":"bef341a7adb10213a7ec7825afeb7d57fbfa7b5f7bdf201204fb0ef62fb9cfa6","state":"FINALITY_PROVIDER_STATUS_ACTIVE","description":{"moniker":"verse2","identity":"","website":"https://verse2.io","security_contact":"ted@verse2.io","details":""},"commission":"0.050000000000000000","active_tvl":0,"active_delegations":0,"type":""},{"btc_pk":"d23c2c25e1fcf8fd1c21b9a402c19e2e309e531e45e92fb1e9805b6056b0cc76","state":"FINALITY_PROVIDER_STATUS_ACTIVE","description":{"moniker":"Babylon Foundation 0","identity":"","website":"https://babylonlabs.io","security_contact":"","details":""},"commission":"0.100000000000000000","active_tvl":0,"active_delegations":0,"type":""},{"btc_pk":"e4889630fa8695dae630c41cd9b85ef165ccc2dc5e5935d5a24393a9defee9ef","state":"FINALITY_PROVIDER_STATUS_ACTIVE","description":{"moniker":"Babylon Foundation 1","identity":"","website":"https://babylonlabs.io","security_contact":"","details":""},"commission":"0.070000000000000000","active_tvl":0,"active_delegations":0,"type":""}],"pagination":{"next_key":""}}`
	assertResponse(t, "/v2/finality-providers", http.StatusOK, contents)
}

func TestV2_NetworkInfo(t *testing.T) {
	cases := []testcase{
		{
			testName:         "ok",
			endpoint:         "/v2/network-info",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"staking_status":{"is_staking_open":true},"params":{"bbn":[{"version":0,"covenant_pks":["49766ccd9e3cd94343e2040474a77fb37cdfd30530d05f9f1e96ae1e2102c86e","76d1ae01f8fb6bf30108731c884cddcf57ef6eef2d9d9559e130894e0e40c62c","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":5000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":2000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":197535,"max_finality_providers":0},{"version":1,"covenant_pks":["09585ab55a971a231c945790a0a81df754e5a07263a5c20829931cc24683bbb7","76d1ae01f8fb6bf30108731c884cddcf57ef6eef2d9d9559e130894e0e40c62c","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":5000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":10000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":198665,"max_finality_providers":0},{"version":2,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":5000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":10000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":200665,"max_finality_providers":0},{"version":3,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":50000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":5000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":215968,"max_finality_providers":0},{"version":4,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":50000000,"min_staking_time_blocks":64000,"max_staking_time_blocks":64000,"slashing_pk_script":"76a914010101010101010101010101010101010101010188ac","min_slashing_tx_fee_sat":1000,"slashing_rate":"0.100000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":5000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1000,"allow_list_expiration_height":26120,"btc_activation_height":220637,"max_finality_providers":0},{"version":5,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":35000000000,"min_staking_time_blocks":10000,"max_staking_time_blocks":64000,"slashing_pk_script":"00145be12624d08a2b424095d7c07221c33450d14bf1","min_slashing_tx_fee_sat":5000,"slashing_rate":"0.050000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":2000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1095000,"allow_list_expiration_height":26124,"btc_activation_height":227174,"max_finality_providers":0},{"version":6,"covenant_pks":["fa9d882d45f4060bdb8042183828cd87544f1ea997380e586cab77d5fd698737","0aee0509b16db71c999238a4827db945526859b13c95487ab46725357c9a9f25","17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e"],"covenant_quorum":6,"min_staking_value_sat":50000,"max_staking_value_sat":35000000000,"min_staking_time_blocks":10000,"max_staking_time_blocks":64000,"slashing_pk_script":"00145be12624d08a2b424095d7c07221c33450d14bf1","min_slashing_tx_fee_sat":6000,"slashing_rate":"0.050000000000000000","unbonding_time_blocks":1008,"unbonding_fee_sat":2000,"min_commission_rate":"0.030000000000000000","delegation_creation_base_gas_fee":1095000,"allow_list_expiration_height":26124,"btc_activation_height":235952,"max_finality_providers":3}],"btc":[{"version":0,"btc_confirmation_depth":10,"checkpoint_finalization_timeout":100}]}}}`,		},
	}

	checkCases(t, cases)
}

func TestV2_Prices(t *testing.T) {
	contents := `{"errorCode":"INTERNAL_SERVICE_ERROR","message":"Internal service error"}`
	assertResponse(t, "/v2/prices", http.StatusInternalServerError, contents)
}

func TestV2_Stats(t *testing.T) {
	contents := `{"data":{"active_tvl":0,"active_delegations":0,"active_finality_providers":3,"total_finality_providers":3,"total_active_tvl":67986511595,"total_active_delegations":417254,"btc_staking_apr":0}}`
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
		{
			testName:         "ok (can_expand=false)",
			endpoint:         "/v2/delegation?staking_tx_hash_hex=d3ef0f9fdbc556bcbc9272d168f4aca8e420880ba624f1691b177a36018c736b",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"can_expand":false,"params_version":1,"staker_btc_pk_hex":"1e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1","finality_provider_btc_pks_hex":["c20acf33c17e5a6c92cced9f1d530cccab7aa3e53400456202f02fac95e9c481"],"delegation_staking":{"staking_tx_hash_hex":"d3ef0f9fdbc556bcbc9272d168f4aca8e420880ba624f1691b177a36018c736b","staking_tx_hex":"02000000000101933cf39b909c9a6b07925ccfadf088559afdac369ac7cec9702e075d3bbeb6090200000000fdffffff0350c30000000000002251201f719f4cacb26c3059abffa5538ee0bbba6d028a9c0a6f754cf46b3f986040f90000000000000000496a4762627434001e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1c20acf33c17e5a6c92cced9f1d530cccab7aa3e53400456202f02fac95e9c481fa00f8fa0300000000002251207754ae229380bdacf33b9011e997655dfb042e01a5ea07ce98b2e716e05b7df601405711f80b8bff9d4751680809b3eea627b042ec82fb12bced7d110f9d9e251fda83e3f3d181de19bc6646efc72ea3ae41bc8043adc20db112f687111368dfb77408080300","staking_output_idx":0,"staking_timelock":64000,"staking_amount":50000,"start_height":198690,"end_height":262690,"bbn_inception_height":37776,"bbn_inception_time":"2025-01-13T09:15:57+04:00","slashing":{"slashing_tx_hex":"","spending_height":0}},"delegation_unbonding":{"unbonding_timelock":1008,"unbonding_tx":"02000000016b738c01367a171b69f124a60b8820e4a8acf468d17292bcbc56c5db9f0fefd30000000000ffffffff01409c000000000000225120dca9ae0b2aa090ebeba2fd05585ffe35bc210bf38388d1139c4bb9aab511f9d500000000","covenant_unbonding_signatures":[{"covenant_btc_pk_hex":"d21faf78c6751a0d38e6bd8028b907ff07e9a869a43fc837d6b3f8dff6119a36","signature_hex":"37fead58bf43ce795e5813735a6809610cb164855223608f7cc85592acc452e924142f045c0d23920a3c4664d37b173d940e3ed8a097a8bb538a5db2313e7382"},{"covenant_btc_pk_hex":"f5199efae3f28bb82476163a7e458c7ad445d9bffb0682d10d3bdb2cb41f8e8e","signature_hex":"c4ff08d4c141d07ab17d87e6948b1ab9e362f074617838c292e59ab56e80cb74a5cc834342cd4f844e2b3f786614f67a2079703747144d3df4ddfdbe2200d221"},{"covenant_btc_pk_hex":"79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0","signature_hex":"c0f59419d1219fbba189538f9318e5f59f6a1d40299883feecbb558d0d68c4fe5589fdb0175eae8d83e5c3b5e94485008dfb68f45feed31960d261926ab84050"},{"covenant_btc_pk_hex":"17921cf156ccb4e73d428f996ed11b245313e37e27c978ac4d2cc21eca4672e4","signature_hex":"37e1a55aba54934294ad1583538e91247a64126a165eb75ec6478bf420543e0e38e43e4cc994fac23b68e5022cf53f6c59a2211a90d39fcd2f0d12efec1a6b0c"},{"covenant_btc_pk_hex":"113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0","signature_hex":"37a1947a5c0c54950a5bbce276f93b5085d782f4efe449bd55832b212d56c98a6f5fce612b6228600794eb5146ade56fd1a0ceaafad3a4ec3c905a5fc3dc88ad"},{"covenant_btc_pk_hex":"3bb93dfc8b61887d771f3630e9a63e97cbafcfcc78556a474df83a31a0ef899c","signature_hex":"f9e0ab8ae40dc4c0eac8ceb3fbcaee483220428426bd46b5f542d98c880b8d00b4fb53713dfe752a4e0733cbc806d5e86758092e7ab678970a7763e267947ba9"},{"covenant_btc_pk_hex":"40afaf47c4ffa56de86410d8e47baa2bb6f04b604f4ea24323737ddc3fe092df","signature_hex":"c39d9e5882102b029b5fe8f9efdd0200b8528744be5b71410b7a8036c98411b817752c6d70de6ffb9c35c5e925174c532308c6595f145287523f28af1d61c987"}],"slashing":{"unbonding_slashing_tx_hex":"","spending_height":0}},"state":"ACTIVE"}}`,
		},
		{
			testName:         "ok (can_expand=true)",
			endpoint:         "/v2/delegation?staking_tx_hash_hex=b4ffb9d0715be3ffe8bbf11c6ee2e3a49931f141ca6c432f8f3d404f67b79ee8",
			expectedHttpCode: http.StatusOK,
			expectedContents: `{"data":{"params_version":0,"staker_btc_pk_hex":"20342b8a35b1d3627cac65e43de5ac484d0e4ed451879debe0583bae543eb25a","finality_provider_btc_pks_hex":["1cf6b50b57c48fac59db84aa04db7bbcce9bdc5704c71c0301b61d072ed06357"],"delegation_staking":{"staking_tx_hash_hex":"b4ffb9d0715be3ffe8bbf11c6ee2e3a49931f141ca6c432f8f3d404f67b79ee8","staking_tx_hex":"010000000196b325275fabe621dbd94cb4e6771f53febad2bb8dc0ff02efbe75786e7702bc0000000000ffffffff0210270000000000002251205123e30fb9e0b3c0c0316f43cb923b5853f1560c36019014a3cae946e0060a4497ff300100000000160014d1cf9898bc4a7e8549682d0da8f199fb8d8a916600000000","staking_timelock":60000,"staking_output_idx":1,"staking_amount":10000,"start_height":257941,"end_height":317941,"bbn_inception_height":129,"bbn_inception_time":"2025-06-17T15:10:24+04:00","slashing":{"slashing_tx_hex":"","spending_height":0}},"delegation_unbonding":{"unbonding_timelock":20,"unbonding_tx":"0200000001e89eb7674f403d8f2f436cca41f13199a4e3e26e1cf1bbe8ffe35b71d0b9ffb40000000000ffffffff01581b000000000000225120cafef41bed50d6b67ea4b7d2d590f815e9c82ceae94b73e1231753cc489e7c7800000000","covenant_unbonding_signatures":[{"covenant_btc_pk_hex":"ffeaec52a9b407b355ef6967a7ffc15fd6c3fe07de2844d61550475e7a5233e5","signature_hex":"8e33e6a71ac70825141579138ed0b8e97c851c646e82dd2be8a2ebbd35d49fc8c9d5750b1d90b1475a4c8d3c4b390eff3873b08950d1f51d3dfaf2541ddb4fe4"},{"covenant_btc_pk_hex":"59d3532148a597a2d05c0395bf5f7176044b1cd312f37701a9b4d0aad70bc5a4","signature_hex":"63da7aa88f61a77cfd583ff85aa27fbdef17ec624bf2502009188f4bb188e2fb6ab8be00d74d0f07f752a9c88b71649bee2d773cab7d53c402f05713cc0f7f45"},{"covenant_btc_pk_hex":"a5c60c2188e833d39d0fa798ab3f69aa12ed3dd2f3bad659effa252782de3c31","signature_hex":"ba0c7ce70dd54b1c1bf9253c0580a9c045f8ef4c23ff72886ce58376b20b5ab4a473793c4b08fd351441ce46c0cabc24ea0a04c48387eb06233e3d54eda60ecf"}],"slashing":{"unbonding_slashing_tx_hex":"","spending_height":0}},"state":"ACTIVE","can_expand":true,"previous_staking_tx_hash_hex":"010000000196b325275fabe621dbd94cb4e6771f53febad2bb8dc0ff02efbe75786e7702bc0000000000ffffffff0210270000000000002251205123e30fb9e0b3c0c0316f43cb923b5853f1560c36019014a3cae946e0060a4497ff300100000000160014d1cf9898bc4a7e8549682d0da8f199fb8d8a916600000001"}}`,
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
			expectedContents: `{"errorCode":"BAD_REQUEST", "message":"staker_pk_hex is required"}`,
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
