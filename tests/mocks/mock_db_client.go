// Code generated by mockery v2.41.0. DO NOT EDIT.

package dbmock

import (
	context "context"

	db "github.com/babylonchain/staking-api-service/internal/db"
	mock "github.com/stretchr/testify/mock"

	model "github.com/babylonchain/staking-api-service/internal/db/model"

	types "github.com/babylonchain/staking-api-service/internal/types"
)

// DBClient is an autogenerated mock type for the DBClient type
type DBClient struct {
	mock.Mock
}

// FindDelegationByTxHashHex provides a mock function with given fields: ctx, txHashHex
func (_m *DBClient) FindDelegationByTxHashHex(ctx context.Context, txHashHex string) (*model.DelegationDocument, error) {
	ret := _m.Called(ctx, txHashHex)

	if len(ret) == 0 {
		panic("no return value specified for FindDelegationByTxHashHex")
	}

	var r0 *model.DelegationDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.DelegationDocument, error)); ok {
		return rf(ctx, txHashHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.DelegationDocument); ok {
		r0 = rf(ctx, txHashHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.DelegationDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, txHashHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindDelegationsByStakerPk provides a mock function with given fields: ctx, stakerPk, paginationToken
func (_m *DBClient) FindDelegationsByStakerPk(ctx context.Context, stakerPk string, paginationToken string) (*db.DbResultMap[model.DelegationDocument], error) {
	ret := _m.Called(ctx, stakerPk, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for FindDelegationsByStakerPk")
	}

	var r0 *db.DbResultMap[model.DelegationDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*db.DbResultMap[model.DelegationDocument], error)); ok {
		return rf(ctx, stakerPk, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *db.DbResultMap[model.DelegationDocument]); ok {
		r0 = rf(ctx, stakerPk, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[model.DelegationDocument])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, stakerPk, paginationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindFinalityProvidersByPkHex provides a mock function with given fields: ctx, pkHex
func (_m *DBClient) FindFinalityProvidersByPkHex(ctx context.Context, pkHex []string) (map[string]model.FinalityProviderDocument, error) {
	ret := _m.Called(ctx, pkHex)

	if len(ret) == 0 {
		panic("no return value specified for FindFinalityProvidersByPkHex")
	}

	var r0 map[string]model.FinalityProviderDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) (map[string]model.FinalityProviderDocument, error)); ok {
		return rf(ctx, pkHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) map[string]model.FinalityProviderDocument); ok {
		r0 = rf(ctx, pkHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]model.FinalityProviderDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, pkHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOrCreateStatsLock provides a mock function with given fields: ctx, stakingTxHashHex, txType
func (_m *DBClient) GetOrCreateStatsLock(ctx context.Context, stakingTxHashHex string, txType string) (*model.StatsLockDocument, error) {
	ret := _m.Called(ctx, stakingTxHashHex, txType)

	if len(ret) == 0 {
		panic("no return value specified for GetOrCreateStatsLock")
	}

	var r0 *model.StatsLockDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.StatsLockDocument, error)); ok {
		return rf(ctx, stakingTxHashHex, txType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.StatsLockDocument); ok {
		r0 = rf(ctx, stakingTxHashHex, txType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.StatsLockDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, stakingTxHashHex, txType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IncrementOverallStats provides a mock function with given fields: ctx, stakingTxHashHex, amount
func (_m *DBClient) IncrementOverallStats(ctx context.Context, stakingTxHashHex string, amount int64) error {
	ret := _m.Called(ctx, stakingTxHashHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for IncrementOverallStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Ping provides a mock function with given fields: ctx
func (_m *DBClient) Ping(ctx context.Context) error {
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

// SaveActiveStakingDelegation provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp
func (_m *DBClient) SaveActiveStakingDelegation(ctx context.Context, stakingTxHashHex string, stakerPkHex string, fpPkHex string, stakingTxHex string, amount uint64, startHeight uint64, timelock uint64, outputIndex uint64, startTimestamp int64) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp)

	if len(ret) == 0 {
		panic("no return value specified for SaveActiveStakingDelegation")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, uint64, uint64, uint64, uint64, int64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveTimeLockExpireCheck provides a mock function with given fields: ctx, stakingTxHashHex, expireHeight, txType
func (_m *DBClient) SaveTimeLockExpireCheck(ctx context.Context, stakingTxHashHex string, expireHeight uint64, txType string) error {
	ret := _m.Called(ctx, stakingTxHashHex, expireHeight, txType)

	if len(ret) == 0 {
		panic("no return value specified for SaveTimeLockExpireCheck")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64, string) error); ok {
		r0 = rf(ctx, stakingTxHashHex, expireHeight, txType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveUnbondingTx provides a mock function with given fields: ctx, stakingTxHashHex, unbondingTxHashHex, txHex, signatureHex
func (_m *DBClient) SaveUnbondingTx(ctx context.Context, stakingTxHashHex string, unbondingTxHashHex string, txHex string, signatureHex string) error {
	ret := _m.Called(ctx, stakingTxHashHex, unbondingTxHashHex, txHex, signatureHex)

	if len(ret) == 0 {
		panic("no return value specified for SaveUnbondingTx")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) error); ok {
		r0 = rf(ctx, stakingTxHashHex, unbondingTxHashHex, txHex, signatureHex)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveUnprocessableMessage provides a mock function with given fields: ctx, messageBody, receipt
func (_m *DBClient) SaveUnprocessableMessage(ctx context.Context, messageBody string, receipt string) error {
	ret := _m.Called(ctx, messageBody, receipt)

	if len(ret) == 0 {
		panic("no return value specified for SaveUnprocessableMessage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, messageBody, receipt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubtractOverallStats provides a mock function with given fields: ctx, stakingTxHashHex, amount
func (_m *DBClient) SubtractOverallStats(ctx context.Context, stakingTxHashHex string, amount int64) error {
	ret := _m.Called(ctx, stakingTxHashHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for SubtractOverallStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TransitionToUnbondedState provides a mock function with given fields: ctx, stakingTxHashHex, eligiblePreviousState
func (_m *DBClient) TransitionToUnbondedState(ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState) error {
	ret := _m.Called(ctx, stakingTxHashHex, eligiblePreviousState)

	if len(ret) == 0 {
		panic("no return value specified for TransitionToUnbondedState")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []types.DelegationState) error); ok {
		r0 = rf(ctx, stakingTxHashHex, eligiblePreviousState)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TransitionToUnbondingState provides a mock function with given fields: ctx, txHashHex, startHeight, timelock, outputIndex, txHex, startTimestamp
func (_m *DBClient) TransitionToUnbondingState(ctx context.Context, txHashHex string, startHeight uint64, timelock uint64, outputIndex uint64, txHex string, startTimestamp int64) error {
	ret := _m.Called(ctx, txHashHex, startHeight, timelock, outputIndex, txHex, startTimestamp)

	if len(ret) == 0 {
		panic("no return value specified for TransitionToUnbondingState")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64, uint64, uint64, string, int64) error); ok {
		r0 = rf(ctx, txHashHex, startHeight, timelock, outputIndex, txHex, startTimestamp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TransitionToWithdrawnState provides a mock function with given fields: ctx, txHashHex
func (_m *DBClient) TransitionToWithdrawnState(ctx context.Context, txHashHex string) error {
	ret := _m.Called(ctx, txHashHex)

	if len(ret) == 0 {
		panic("no return value specified for TransitionToWithdrawnState")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, txHashHex)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDBClient creates a new instance of DBClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDBClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *DBClient {
	mock := &DBClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
