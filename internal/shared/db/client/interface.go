package dbclient

import (
	"context"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
)

//go:generate mockery --name=DBClient --output=../../../../tests/mocks --outpkg=mocks --filename=mock_db_client.go
type DBClient interface {
	Ping(ctx context.Context) error
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
	) ([]*dbmodel.PkAddressMapping, error)
	// FindPkMappingsByNativeSegwitAddress finds the PK address mappings by native
	// segwit address. The returned slice addressMapping will only contain
	// documents for addresses that were found in the database.
	// If some addresses do not have a matching document, those addresses will
	// simply be absent from the result.
	FindPkMappingsByNativeSegwitAddress(
		ctx context.Context, nativeSegwitAddresses []string,
	) ([]*dbmodel.PkAddressMapping, error)
	SaveUnprocessableMessage(ctx context.Context, messageBody, receipt string) error
	FindUnprocessableMessages(ctx context.Context) ([]dbmodel.UnprocessableMessageDocument, error)
	DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error

	// GetLatestBtcPrice fetches the BTC price from the database.
	GetLatestBtcPrice(ctx context.Context) (*dbmodel.BtcPrice, error)
	// SetBtcPrice sets the latest BTC price in the database.
	SetBtcPrice(ctx context.Context, price float64) error
}
