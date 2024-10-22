package v2dbclient

import (
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
)

type V2DBClientInterface interface {
	dbclient.DBClientInterface
}
