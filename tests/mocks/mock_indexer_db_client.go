// Code generated by mockery v2.41.0. DO NOT EDIT.

package mocks

import (
	context "context"

	indexerdbclient "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/client"
	db "github.com/babylonlabs-io/staking-api-service/internal/shared/db"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"

	mock "github.com/stretchr/testify/mock"
)

// IndexerDBClient is an autogenerated mock type for the IndexerDBClient type
type IndexerDBClient struct {
	mock.Mock
}

// CheckDelegationExistByStakerPk provides a mock function with given fields: ctx, address, extraFilter
func (_m *IndexerDBClient) CheckDelegationExistByStakerPk(ctx context.Context, address string, extraFilter *indexerdbclient.DelegationFilter) (bool, error) {
	ret := _m.Called(ctx, address, extraFilter)

	if len(ret) == 0 {
		panic("no return value specified for CheckDelegationExistByStakerPk")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *indexerdbclient.DelegationFilter) (bool, error)); ok {
		return rf(ctx, address, extraFilter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *indexerdbclient.DelegationFilter) bool); ok {
		r0 = rf(ctx, address, extraFilter)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *indexerdbclient.DelegationFilter) error); ok {
		r1 = rf(ctx, address, extraFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBbnStakingParams provides a mock function with given fields: ctx
func (_m *IndexerDBClient) GetBbnStakingParams(ctx context.Context) ([]*indexertypes.BbnStakingParams, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetBbnStakingParams")
	}

	var r0 []*indexertypes.BbnStakingParams
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*indexertypes.BbnStakingParams, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*indexertypes.BbnStakingParams); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*indexertypes.BbnStakingParams)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBtcCheckpointParams provides a mock function with given fields: ctx
func (_m *IndexerDBClient) GetBtcCheckpointParams(ctx context.Context) ([]*indexertypes.BtcCheckpointParams, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetBtcCheckpointParams")
	}

	var r0 []*indexertypes.BtcCheckpointParams
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*indexertypes.BtcCheckpointParams, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*indexertypes.BtcCheckpointParams); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*indexertypes.BtcCheckpointParams)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDelegation provides a mock function with given fields: ctx, stakingTxHashHex
func (_m *IndexerDBClient) GetDelegation(ctx context.Context, stakingTxHashHex string) (*indexerdbmodel.IndexerDelegationDetails, error) {
	ret := _m.Called(ctx, stakingTxHashHex)

	if len(ret) == 0 {
		panic("no return value specified for GetDelegation")
	}

	var r0 *indexerdbmodel.IndexerDelegationDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*indexerdbmodel.IndexerDelegationDetails, error)); ok {
		return rf(ctx, stakingTxHashHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *indexerdbmodel.IndexerDelegationDetails); ok {
		r0 = rf(ctx, stakingTxHashHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*indexerdbmodel.IndexerDelegationDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, stakingTxHashHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDelegations provides a mock function with given fields: ctx, stakerPKHex, stakerBabylonAddress, paginationToken
func (_m *IndexerDBClient) GetDelegations(ctx context.Context, stakerPKHex string, stakerBabylonAddress *string, paginationToken string) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error) {
	ret := _m.Called(ctx, stakerPKHex, stakerBabylonAddress, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for GetDelegations")
	}

	var r0 *db.DbResultMap[indexerdbmodel.IndexerDelegationDetails]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *string, string) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error)); ok {
		return rf(ctx, stakerPKHex, stakerBabylonAddress, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *string, string) *db.DbResultMap[indexerdbmodel.IndexerDelegationDetails]); ok {
		r0 = rf(ctx, stakerPKHex, stakerBabylonAddress, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *string, string) error); ok {
		r1 = rf(ctx, stakerPKHex, stakerBabylonAddress, paginationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDelegationsInStates provides a mock function with given fields: ctx, stakerPKHex, stakerBabylonAddress, states
func (_m *IndexerDBClient) GetDelegationsInStates(ctx context.Context, stakerPKHex string, stakerBabylonAddress *string, states []indexertypes.DelegationState) ([]indexerdbmodel.IndexerDelegationDetails, error) {
	ret := _m.Called(ctx, stakerPKHex, stakerBabylonAddress, states)

	if len(ret) == 0 {
		panic("no return value specified for GetDelegationsInStates")
	}

	var r0 []indexerdbmodel.IndexerDelegationDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *string, []indexertypes.DelegationState) ([]indexerdbmodel.IndexerDelegationDetails, error)); ok {
		return rf(ctx, stakerPKHex, stakerBabylonAddress, states)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *string, []indexertypes.DelegationState) []indexerdbmodel.IndexerDelegationDetails); ok {
		r0 = rf(ctx, stakerPKHex, stakerBabylonAddress, states)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]indexerdbmodel.IndexerDelegationDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *string, []indexertypes.DelegationState) error); ok {
		r1 = rf(ctx, stakerPKHex, stakerBabylonAddress, states)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFinalityProviderByPk provides a mock function with given fields: ctx, fpPk
func (_m *IndexerDBClient) GetFinalityProviderByPk(ctx context.Context, fpPk string) (*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	ret := _m.Called(ctx, fpPk)

	if len(ret) == 0 {
		panic("no return value specified for GetFinalityProviderByPk")
	}

	var r0 *indexerdbmodel.IndexerFinalityProviderDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*indexerdbmodel.IndexerFinalityProviderDetails, error)); ok {
		return rf(ctx, fpPk)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *indexerdbmodel.IndexerFinalityProviderDetails); ok {
		r0 = rf(ctx, fpPk)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*indexerdbmodel.IndexerFinalityProviderDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, fpPk)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFinalityProviders provides a mock function with given fields: ctx
func (_m *IndexerDBClient) GetFinalityProviders(ctx context.Context) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetFinalityProviders")
	}

	var r0 []*indexerdbmodel.IndexerFinalityProviderDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*indexerdbmodel.IndexerFinalityProviderDetails); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*indexerdbmodel.IndexerFinalityProviderDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLastProcessedBbnHeight provides a mock function with given fields: ctx
func (_m *IndexerDBClient) GetLastProcessedBbnHeight(ctx context.Context) (uint64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLastProcessedBbnHeight")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields: ctx
func (_m *IndexerDBClient) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewIndexerDBClient creates a new instance of IndexerDBClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIndexerDBClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *IndexerDBClient {
	mock := &IndexerDBClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
