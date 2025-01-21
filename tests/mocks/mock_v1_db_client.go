// Code generated by mockery v2.51.0. DO NOT EDIT.

package mocks

import (
	context "context"

	db "github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"

	mock "github.com/stretchr/testify/mock"

	types "github.com/babylonlabs-io/staking-api-service/internal/shared/types"

	v1dbclient "github.com/babylonlabs-io/staking-api-service/internal/v1/db/client"

	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
)

// V1DBClient is an autogenerated mock type for the V1DBClient type
type V1DBClient struct {
	mock.Mock
}

// CheckDelegationExistByStakerPk provides a mock function with given fields: ctx, address, extraFilter
func (_m *V1DBClient) CheckDelegationExistByStakerPk(ctx context.Context, address string, extraFilter *v1dbclient.DelegationFilter) (bool, error) {
	ret := _m.Called(ctx, address, extraFilter)

	if len(ret) == 0 {
		panic("no return value specified for CheckDelegationExistByStakerPk")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1dbclient.DelegationFilter) (bool, error)); ok {
		return rf(ctx, address, extraFilter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1dbclient.DelegationFilter) bool); ok {
		r0 = rf(ctx, address, extraFilter)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *v1dbclient.DelegationFilter) error); ok {
		r1 = rf(ctx, address, extraFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteUnprocessableMessage provides a mock function with given fields: ctx, Receipt
func (_m *V1DBClient) DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error {
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
func (_m *V1DBClient) FindDelegationByTxHashHex(ctx context.Context, txHashHex string) (*v1dbmodel.DelegationDocument, error) {
	ret := _m.Called(ctx, txHashHex)

	if len(ret) == 0 {
		panic("no return value specified for FindDelegationByTxHashHex")
	}

	var r0 *v1dbmodel.DelegationDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1dbmodel.DelegationDocument, error)); ok {
		return rf(ctx, txHashHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1dbmodel.DelegationDocument); ok {
		r0 = rf(ctx, txHashHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1dbmodel.DelegationDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, txHashHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindDelegationsByStakerPk provides a mock function with given fields: ctx, stakerPk, extraFilter, paginationToken
func (_m *V1DBClient) FindDelegationsByStakerPk(ctx context.Context, stakerPk string, extraFilter *v1dbclient.DelegationFilter, paginationToken string) (*db.DbResultMap[v1dbmodel.DelegationDocument], error) {
	ret := _m.Called(ctx, stakerPk, extraFilter, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for FindDelegationsByStakerPk")
	}

	var r0 *db.DbResultMap[v1dbmodel.DelegationDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1dbclient.DelegationFilter, string) (*db.DbResultMap[v1dbmodel.DelegationDocument], error)); ok {
		return rf(ctx, stakerPk, extraFilter, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1dbclient.DelegationFilter, string) *db.DbResultMap[v1dbmodel.DelegationDocument]); ok {
		r0 = rf(ctx, stakerPk, extraFilter, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[v1dbmodel.DelegationDocument])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *v1dbclient.DelegationFilter, string) error); ok {
		r1 = rf(ctx, stakerPk, extraFilter, paginationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindFinalityProviderStats provides a mock function with given fields: ctx, paginationToken
func (_m *V1DBClient) FindFinalityProviderStats(ctx context.Context, paginationToken string) (*db.DbResultMap[*v1dbmodel.FinalityProviderStatsDocument], error) {
	ret := _m.Called(ctx, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for FindFinalityProviderStats")
	}

	var r0 *db.DbResultMap[*v1dbmodel.FinalityProviderStatsDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*db.DbResultMap[*v1dbmodel.FinalityProviderStatsDocument], error)); ok {
		return rf(ctx, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *db.DbResultMap[*v1dbmodel.FinalityProviderStatsDocument]); ok {
		r0 = rf(ctx, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[*v1dbmodel.FinalityProviderStatsDocument])
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
func (_m *V1DBClient) FindFinalityProviderStatsByFinalityProviderPkHex(ctx context.Context, finalityProviderPkHex []string) ([]*v1dbmodel.FinalityProviderStatsDocument, error) {
	ret := _m.Called(ctx, finalityProviderPkHex)

	if len(ret) == 0 {
		panic("no return value specified for FindFinalityProviderStatsByFinalityProviderPkHex")
	}

	var r0 []*v1dbmodel.FinalityProviderStatsDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*v1dbmodel.FinalityProviderStatsDocument, error)); ok {
		return rf(ctx, finalityProviderPkHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*v1dbmodel.FinalityProviderStatsDocument); ok {
		r0 = rf(ctx, finalityProviderPkHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1dbmodel.FinalityProviderStatsDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, finalityProviderPkHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindPkMappingsByNativeSegwitAddress provides a mock function with given fields: ctx, nativeSegwitAddresses
func (_m *V1DBClient) FindPkMappingsByNativeSegwitAddress(ctx context.Context, nativeSegwitAddresses []string) ([]*dbmodel.PkAddressMapping, error) {
	ret := _m.Called(ctx, nativeSegwitAddresses)

	if len(ret) == 0 {
		panic("no return value specified for FindPkMappingsByNativeSegwitAddress")
	}

	var r0 []*dbmodel.PkAddressMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*dbmodel.PkAddressMapping, error)); ok {
		return rf(ctx, nativeSegwitAddresses)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*dbmodel.PkAddressMapping); ok {
		r0 = rf(ctx, nativeSegwitAddresses)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*dbmodel.PkAddressMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, nativeSegwitAddresses)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindPkMappingsByTaprootAddress provides a mock function with given fields: ctx, taprootAddresses
func (_m *V1DBClient) FindPkMappingsByTaprootAddress(ctx context.Context, taprootAddresses []string) ([]*dbmodel.PkAddressMapping, error) {
	ret := _m.Called(ctx, taprootAddresses)

	if len(ret) == 0 {
		panic("no return value specified for FindPkMappingsByTaprootAddress")
	}

	var r0 []*dbmodel.PkAddressMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*dbmodel.PkAddressMapping, error)); ok {
		return rf(ctx, taprootAddresses)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*dbmodel.PkAddressMapping); ok {
		r0 = rf(ctx, taprootAddresses)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*dbmodel.PkAddressMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, taprootAddresses)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindTopStakersByTvl provides a mock function with given fields: ctx, paginationToken
func (_m *V1DBClient) FindTopStakersByTvl(ctx context.Context, paginationToken string) (*db.DbResultMap[*v1dbmodel.StakerStatsDocument], error) {
	ret := _m.Called(ctx, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for FindTopStakersByTvl")
	}

	var r0 *db.DbResultMap[*v1dbmodel.StakerStatsDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*db.DbResultMap[*v1dbmodel.StakerStatsDocument], error)); ok {
		return rf(ctx, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *db.DbResultMap[*v1dbmodel.StakerStatsDocument]); ok {
		r0 = rf(ctx, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[*v1dbmodel.StakerStatsDocument])
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
func (_m *V1DBClient) FindUnprocessableMessages(ctx context.Context) ([]dbmodel.UnprocessableMessageDocument, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for FindUnprocessableMessages")
	}

	var r0 []dbmodel.UnprocessableMessageDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]dbmodel.UnprocessableMessageDocument, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []dbmodel.UnprocessableMessageDocument); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]dbmodel.UnprocessableMessageDocument)
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
func (_m *V1DBClient) GetLatestBtcInfo(ctx context.Context) (*v1dbmodel.BtcInfo, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestBtcInfo")
	}

	var r0 *v1dbmodel.BtcInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*v1dbmodel.BtcInfo, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *v1dbmodel.BtcInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1dbmodel.BtcInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatestBtcPrice provides a mock function with given fields: ctx
func (_m *V1DBClient) GetLatestBtcPrice(ctx context.Context) (*dbmodel.BtcPrice, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestBtcPrice")
	}

	var r0 *dbmodel.BtcPrice
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*dbmodel.BtcPrice, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *dbmodel.BtcPrice); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dbmodel.BtcPrice)
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
func (_m *V1DBClient) GetOrCreateStatsLock(ctx context.Context, stakingTxHashHex string, state string) (*v1dbmodel.StatsLockDocument, error) {
	ret := _m.Called(ctx, stakingTxHashHex, state)

	if len(ret) == 0 {
		panic("no return value specified for GetOrCreateStatsLock")
	}

	var r0 *v1dbmodel.StatsLockDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*v1dbmodel.StatsLockDocument, error)); ok {
		return rf(ctx, stakingTxHashHex, state)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1dbmodel.StatsLockDocument); ok {
		r0 = rf(ctx, stakingTxHashHex, state)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1dbmodel.StatsLockDocument)
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
func (_m *V1DBClient) GetOverallStats(ctx context.Context) (*v1dbmodel.OverallStatsDocument, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetOverallStats")
	}

	var r0 *v1dbmodel.OverallStatsDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*v1dbmodel.OverallStatsDocument, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *v1dbmodel.OverallStatsDocument); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1dbmodel.OverallStatsDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStakerStats provides a mock function with given fields: ctx, stakerPkHex
func (_m *V1DBClient) GetStakerStats(ctx context.Context, stakerPkHex string) (*v1dbmodel.StakerStatsDocument, error) {
	ret := _m.Called(ctx, stakerPkHex)

	if len(ret) == 0 {
		panic("no return value specified for GetStakerStats")
	}

	var r0 *v1dbmodel.StakerStatsDocument
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1dbmodel.StakerStatsDocument, error)); ok {
		return rf(ctx, stakerPkHex)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1dbmodel.StakerStatsDocument); ok {
		r0 = rf(ctx, stakerPkHex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1dbmodel.StakerStatsDocument)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, stakerPkHex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IncrementFinalityProviderStats provides a mock function with given fields: ctx, stakingTxHashHex, fpPkHex, amount
func (_m *V1DBClient) IncrementFinalityProviderStats(ctx context.Context, stakingTxHashHex string, fpPkHex string, amount uint64) error {
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
func (_m *V1DBClient) IncrementOverallStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
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
func (_m *V1DBClient) IncrementStakerStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
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
func (_m *V1DBClient) InsertPkAddressMappings(ctx context.Context, stakerPkHex string, taproot string, nativeSigwitOdd string, nativeSigwitEven string) error {
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
func (_m *V1DBClient) Ping(ctx context.Context) error {
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

// SaveActiveStakingDelegation provides a mock function with given fields: ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp, isOverflow
func (_m *V1DBClient) SaveActiveStakingDelegation(ctx context.Context, stakingTxHashHex string, stakerPkHex string, fpPkHex string, stakingTxHex string, amount uint64, startHeight uint64, timelock uint64, outputIndex uint64, startTimestamp int64, isOverflow bool) error {
	ret := _m.Called(ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp, isOverflow)

	if len(ret) == 0 {
		panic("no return value specified for SaveActiveStakingDelegation")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, uint64, uint64, uint64, uint64, int64, bool) error); ok {
		r0 = rf(ctx, stakingTxHashHex, stakerPkHex, fpPkHex, stakingTxHex, amount, startHeight, timelock, outputIndex, startTimestamp, isOverflow)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveTimeLockExpireCheck provides a mock function with given fields: ctx, stakingTxHashHex, expireHeight, txType
func (_m *V1DBClient) SaveTimeLockExpireCheck(ctx context.Context, stakingTxHashHex string, expireHeight uint64, txType string) error {
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
func (_m *V1DBClient) SaveUnbondingTx(ctx context.Context, stakingTxHashHex string, unbondingTxHashHex string, txHex string, signatureHex string) error {
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
func (_m *V1DBClient) SaveUnprocessableMessage(ctx context.Context, messageBody string, receipt string) error {
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

// ScanDelegationsPaginated provides a mock function with given fields: ctx, paginationToken
func (_m *V1DBClient) ScanDelegationsPaginated(ctx context.Context, paginationToken string) (*db.DbResultMap[v1dbmodel.DelegationDocument], error) {
	ret := _m.Called(ctx, paginationToken)

	if len(ret) == 0 {
		panic("no return value specified for ScanDelegationsPaginated")
	}

	var r0 *db.DbResultMap[v1dbmodel.DelegationDocument]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*db.DbResultMap[v1dbmodel.DelegationDocument], error)); ok {
		return rf(ctx, paginationToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *db.DbResultMap[v1dbmodel.DelegationDocument]); ok {
		r0 = rf(ctx, paginationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.DbResultMap[v1dbmodel.DelegationDocument])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, paginationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetBtcPrice provides a mock function with given fields: ctx, price
func (_m *V1DBClient) SetBtcPrice(ctx context.Context, price float64) error {
	ret := _m.Called(ctx, price)

	if len(ret) == 0 {
		panic("no return value specified for SetBtcPrice")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, float64) error); ok {
		r0 = rf(ctx, price)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubtractFinalityProviderStats provides a mock function with given fields: ctx, stakingTxHashHex, fpPkHex, amount
func (_m *V1DBClient) SubtractFinalityProviderStats(ctx context.Context, stakingTxHashHex string, fpPkHex string, amount uint64) error {
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
func (_m *V1DBClient) SubtractOverallStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
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
func (_m *V1DBClient) SubtractStakerStats(ctx context.Context, stakingTxHashHex string, stakerPkHex string, amount uint64) error {
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

// TransitionToTransitionedState provides a mock function with given fields: ctx, stakingTxHashHex
func (_m *V1DBClient) TransitionToTransitionedState(ctx context.Context, stakingTxHashHex string) error {
	ret := _m.Called(ctx, stakingTxHashHex)

	if len(ret) == 0 {
		panic("no return value specified for TransitionToTransitionedState")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, stakingTxHashHex)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TransitionToUnbondedState provides a mock function with given fields: ctx, stakingTxHashHex, eligiblePreviousState
func (_m *V1DBClient) TransitionToUnbondedState(ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState) error {
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
func (_m *V1DBClient) TransitionToUnbondingState(ctx context.Context, txHashHex string, startHeight uint64, timelock uint64, outputIndex uint64, txHex string, startTimestamp int64) error {
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
func (_m *V1DBClient) TransitionToWithdrawnState(ctx context.Context, txHashHex string) error {
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
func (_m *V1DBClient) UpsertLatestBtcInfo(ctx context.Context, height uint64, confirmedTvl uint64, unconfirmedTvl uint64) error {
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

// NewV1DBClient creates a new instance of V1DBClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewV1DBClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *V1DBClient {
	mock := &V1DBClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
