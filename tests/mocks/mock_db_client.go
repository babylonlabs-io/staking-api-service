// Code generated by mockery v2.41.0. DO NOT EDIT.

package mocks

import (
	context "context"

	db "github.com/babylonlabs-io/staking-api-service/internal/db"
	mock "github.com/stretchr/testify/mock"

	model "github.com/babylonlabs-io/staking-api-service/internal/db/model"

	types "github.com/babylonlabs-io/staking-api-service/internal/types"
)

// DBClient is an autogenerated mock type for the DBClient type
type DBClient struct {
	mock.Mock
}

// CheckDelegationExistByStakerTaprootAddress provides a mock function with given fields: ctx, address, extraFilter
func (_m *DBClient) CheckDelegationExistByStakerTaprootAddress(ctx context.Context, address string, extraFilter *db.DelegationFilter) (bool, error) {
	ret := _m.Called(ctx, address, extraFilter)

	if len(ret) == 0 {
		panic("no return value specified for CheckDelegationExistByStakerTaprootAddress")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *db.DelegationFilter) (bool, error)); ok {
		return rf(ctx, address, extraFilter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *db.DelegationFilter) bool); ok {
		r0 = rf(ctx, address, extraFilter)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *db.DelegationFilter) error); ok {
		r1 = rf(ctx, address, extraFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteUnprocessableMessage provides a mock function with given fields: ctx, Receipt
func (_m *DBClient) DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error {
	ret := _m.Called(ctx, Receipt)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUnprocessableMessage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) error); ok {
		r0 = rf(ctx, Receipt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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

// FindFinalityProviderStats provides a mock function with given fields: ctx, paginationToken
func (_m *DBClient) FindFinalityProviderStats(ctx context.Context, paginationToken string) (*db.DbResultMap[*model.FinalityProviderStatsDocument], error) {
	ret := _m.Called(ctx, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for FindFinalityProviderStats")
	}

	var r0 *db.DbResultMap[*model.FinalityProviderStatsDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*db.DbResultMap[*model.FinalityProviderStatsDocument], error)); ok {
		return rf(ctx, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *db.DbResultMap[*model.FinalityProviderStatsDocument]); ok {
		r0 = rf(ctx, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[*model.FinalityProviderStatsDocument])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, paginationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindFinalityProviderStatsByFinalityProviderPkHex provides a mock function with given fields: ctx, finalityProviderPkHex
func (_m *DBClient) FindFinalityProviderStatsByFinalityProviderPkHex(ctx context.Context, finalityProviderPkHex []string) ([]*model.FinalityProviderStatsDocument, error) {
	ret := _m.Called(ctx, finalityProviderPkHex)

	if len(ret) == 0 {
		panic("no return value specified for FindFinalityProviderStatsByFinalityProviderPkHex")
	}

	var r0 []*model.FinalityProviderStatsDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*model.FinalityProviderStatsDocument, error)); ok {
		return rf(ctx, finalityProviderPkHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.FinalityProviderStatsDocument); ok {
		r0 = rf(ctx, finalityProviderPkHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FinalityProviderStatsDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, finalityProviderPkHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindTopStakersByTvl provides a mock function with given fields: ctx, paginationToken
func (_m *DBClient) FindTopStakersByTvl(ctx context.Context, paginationToken string) (*db.DbResultMap[*model.StakerStatsDocument], error) {
	ret := _m.Called(ctx, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for FindTopStakersByTvl")
	}

	var r0 *db.DbResultMap[*model.StakerStatsDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*db.DbResultMap[*model.StakerStatsDocument], error)); ok {
		return rf(ctx, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *db.DbResultMap[*model.StakerStatsDocument]); ok {
		r0 = rf(ctx, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[*model.StakerStatsDocument])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, paginationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindUnprocessableMessages provides a mock function with given fields: ctx
func (_m *DBClient) FindUnprocessableMessages(ctx context.Context) ([]model.UnprocessableMessageDocument, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for FindUnprocessableMessages")
	}

	var r0 []model.UnprocessableMessageDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]model.UnprocessableMessageDocument, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []model.UnprocessableMessageDocument); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.UnprocessableMessageDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatestBtcInfo provides a mock function with given fields: ctx
func (_m *DBClient) GetLatestBtcInfo(ctx context.Context) (*model.BtcInfo, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestBtcInfo")
	}

	var r0 *model.BtcInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*model.BtcInfo, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *model.BtcInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BtcInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOrCreateStatsLock provides a mock function with given fields: ctx, stakingTxHashHex, state
func (_m *DBClient) GetOrCreateStatsLock(ctx context.Context, stakingTxHashHex string, state string) (*model.StatsLockDocument, error) {
	ret := _m.Called(ctx, stakingTxHashHex, state)

	if len(ret) == 0 {
		panic("no return value specified for GetOrCreateStatsLock")
	}

	var r0 *model.StatsLockDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.StatsLockDocument, error)); ok {
		return rf(ctx, stakingTxHashHex, state)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.StatsLockDocument); ok {
		r0 = rf(ctx, stakingTxHashHex, state)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.StatsLockDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, stakingTxHashHex, state)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOverallStats provides a mock function with given fields: ctx
func (_m *DBClient) GetOverallStats(ctx context.Context) (*model.OverallStatsDocument, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetOverallStats")
	}

	var r0 *model.OverallStatsDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*model.OverallStatsDocument, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *model.OverallStatsDocument); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.OverallStatsDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IncrementFinalityProviderStats provides a mock function with given fields: ctx, stakingTxHashHex, fpPkHex, amount
func (_m *DBClient) IncrementFinalityProviderStats(ctx context.Context, stakingTxHashHex string, fpPkHex string, amount uint64) error {
	ret := _m.Called(ctx, stakingTxHashHex, fpPkHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for IncrementFinalityProviderStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, fpPkHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IncrementOverallStats provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, amount
func (_m *DBClient) IncrementOverallStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for IncrementOverallStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IncrementStakerStats provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, amount
func (_m *DBClient) IncrementStakerStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for IncrementStakerStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InsertPkAddressMappings provides a mock function with given fields: ctx, stakerPkHex, taproot, nativeSigwitOdd, nativeSigwitEven
func (_m *DBClient) InsertPkAddressMappings(ctx context.Context, stakerPkHex string, taproot string, nativeSigwitOdd string, nativeSigwitEven string) error {
	ret := _m.Called(ctx, stakerPkHex, taproot, nativeSigwitOdd, nativeSigwitEven)

	if len(ret) == 0 {
		panic("no return value specified for InsertPkAddressMappings")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) error); ok {
		r0 = rf(ctx, stakerPkHex, taproot, nativeSigwitOdd, nativeSigwitEven)
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

// SaveActiveStakingDelegation provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp, isOverflow, stakerTaprootAddress
func (_m *DBClient) SaveActiveStakingDelegation(ctx context.Context, stakingTxHashHex string, stakerPkHex string, fpPkHex string, stakingTxHex string, amount uint64, startHeight uint64, timelock uint64, outputIndex uint64, startTimestamp int64, isOverflow bool, stakerTaprootAddress string) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp, isOverflow, stakerTaprootAddress)

	if len(ret) == 0 {
		panic("no return value specified for SaveActiveStakingDelegation")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, uint64, uint64, uint64, uint64, int64, bool, string) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp, isOverflow, stakerTaprootAddress)
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

// SubtractFinalityProviderStats provides a mock function with given fields: ctx, stakingTxHashHex, fpPkHex, amount
func (_m *DBClient) SubtractFinalityProviderStats(ctx context.Context, stakingTxHashHex string, fpPkHex string, amount uint64) error {
	ret := _m.Called(ctx, stakingTxHashHex, fpPkHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for SubtractFinalityProviderStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, fpPkHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubtractOverallStats provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, amount
func (_m *DBClient) SubtractOverallStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for SubtractOverallStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubtractStakerStats provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, amount
func (_m *DBClient) SubtractStakerStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, amount)

	if len(ret) == 0 {
		panic("no return value specified for SubtractStakerStats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint64) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, amount)
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

// UpsertLatestBtcInfo provides a mock function with given fields: ctx, height, confirmedTvl, unconfirmedTvl
func (_m *DBClient) UpsertLatestBtcInfo(ctx context.Context, height uint64, confirmedTvl uint64, unconfirmedTvl uint64) error {
	ret := _m.Called(ctx, height, confirmedTvl, unconfirmedTvl)

	if len(ret) == 0 {
		panic("no return value specified for UpsertLatestBtcInfo")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, uint64, uint64) error); ok {
		r0 = rf(ctx, height, confirmedTvl, unconfirmedTvl)
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
