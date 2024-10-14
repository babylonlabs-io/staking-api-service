package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
)

type BaseDBClient interface {
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
	) ([]*model.PkAddressMapping, error)
	// FindPkMappingsByNativeSegwitAddress finds the PK address mappings by native
	// segwit address. The returned slice addressMapping will only contain
	// documents for addresses that were found in the database.
	// If some addresses do not have a matching document, those addresses will
	// simply be absent from the result.
	FindPkMappingsByNativeSegwitAddress(
		ctx context.Context, nativeSegwitAddresses []string,
	) ([]*model.PkAddressMapping, error)
	SaveUnprocessableMessage(ctx context.Context, messageBody, receipt string) error
	FindUnprocessableMessages(ctx context.Context) ([]model.UnprocessableMessageDocument, error)
	DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error
}
