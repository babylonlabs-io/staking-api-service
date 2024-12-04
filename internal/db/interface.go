package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type DBClient interface {
	Ping(ctx context.Context) error
	SaveActiveStakingDelegation(
		ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string,
		stakingTxHex string, amount, startHeight, timelock, outputIndex uint64,
		startTimestamp int64, isOverflow bool,
	) error
	// FindDelegationsByStakerPk finds all delegations by the staker's public key.
	// The extraFilter parameter can be used to filter the results by the delegation's
	// properties. The paginationToken parameter is used to fetch the next page of results.
	// If the paginationToken is empty, the first page of results will be fetched.
	// The returned DbResultMap will contain the next pagination token if there are more
	// results to fetch.
	FindDelegationsByStakerPk(
		ctx context.Context, stakerPk string,
		extraFilter *DelegationFilter, paginationToken string,
	) (*DbResultMap[model.DelegationDocument], error)
	SaveUnbondingTx(
		ctx context.Context, stakingTxHashHex, unbondingTxHashHex, txHex, signatureHex string,
	) error
	FindDelegationByTxHashHex(ctx context.Context, txHashHex string) (*model.DelegationDocument, error)
	SaveTimeLockExpireCheck(ctx context.Context, stakingTxHashHex string, expireHeight uint64, txType string) error
	SaveUnprocessableMessage(ctx context.Context, messageBody, receipt string) error
	FindUnprocessableMessages(ctx context.Context) ([]model.UnprocessableMessageDocument, error)
	DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error
	TransitionToUnbondedState(
		ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState,
	) error
	TransitionToUnbondingState(
		ctx context.Context, txHashHex string, startHeight, timelock, outputIndex uint64, txHex string, startTimestamp int64,
	) error
	TransitionToWithdrawnState(ctx context.Context, txHashHex string) error
	GetOrCreateStatsLock(
		ctx context.Context, stakingTxHashHex string, state string,
	) (*model.StatsLockDocument, error)
	SubtractOverallStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	IncrementOverallStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	GetOverallStats(ctx context.Context) (*model.OverallStatsDocument, error)
	IncrementFinalityProviderStats(
		ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
	) error
	SubtractFinalityProviderStats(
		ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
	) error
	FindFinalityProviderStats(ctx context.Context, paginationToken string) (*DbResultMap[*model.FinalityProviderStatsDocument], error)
	FindFinalityProviderStatsByFinalityProviderPkHex(
		ctx context.Context, finalityProviderPkHex []string,
	) ([]*model.FinalityProviderStatsDocument, error)
	IncrementStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	SubtractStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	FindTopStakersByTvl(ctx context.Context, paginationToken string) (*DbResultMap[*model.StakerStatsDocument], error)
	// GetStakerStats fetches the staker stats by the staker's public key.
	GetStakerStats(
		ctx context.Context, stakerPkHex string,
	) (*model.StakerStatsDocument, error)
	UpsertLatestBtcInfo(
		ctx context.Context, height uint64, confirmedTvl uint64, unconfirmedTvl uint64,
	) error
	GetLatestBtcInfo(ctx context.Context) (*model.BtcInfo, error)
	CheckDelegationExistByStakerPk(
		ctx context.Context, address string, extraFilter *DelegationFilter,
	) (bool, error)
	// InsertPkAddressMappings inserts the btc public key and
	// its corresponding btc addresses into the database.
	InsertPkAddressMappings(
		ctx context.Context, stakerPkHex, taproot, nativeSigwitOdd, nativeSigwitEven string,
	) error
	// FindPkMappingsByTaprootAddress finds the PK address mappings by taproot address.
	// The returned slice addressMapping will only contain documents for addresses
	// that were found in the database. If some addresses do not have a matching
	// document, those addresses will simply be absent from the result.
	FindPkMappingsByTaprootAddress(
		ctx context.Context, taprootAddresses []string,
	) ([]*model.PkAddressMapping, error)
	// FindPkMappingsByNativeSegwitAddress finds the PK address mappings by native
	// segwit address. The returned slice addressMapping will only contain
	// documents for addresses that were found in the database.
	// If some addresses do not have a matching document, those addresses will
	// simply be absent from the result.
	FindPkMappingsByNativeSegwitAddress(
		ctx context.Context, nativeSegwitAddresses []string,
	) ([]*model.PkAddressMapping, error)
	// ScanDelegationsPaginated scans the delegation collection in a paginated way
	// without applying any filters or sorting, ensuring that all existing items
	// are eventually fetched.
	ScanDelegationsPaginated(
		ctx context.Context,
		paginationToken string,
	) (*DbResultMap[model.DelegationDocument], error)
	// SaveTermsAcceptance saves the acceptance of the terms of service of the public key
	SaveTermsAcceptance(ctx context.Context, termsAcceptance *model.TermsAcceptance) error
	// GetLatestBtcPrice fetches the BTC price from the database.
	GetLatestBtcPrice(ctx context.Context) (*model.BtcPrice, error)
	// SetBtcPrice sets the latest BTC price in the database.
	SetBtcPrice(ctx context.Context, price float64) error
}

type DelegationFilter struct {
	AfterTimestamp int64
	States         []types.DelegationState
}
