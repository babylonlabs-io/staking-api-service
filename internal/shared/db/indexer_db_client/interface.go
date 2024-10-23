package indexerdbclient

import (
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
)

type IndexerDBClient interface {
	dbclient.DBClient
}
