package v2db

import (
	"github.com/babylonlabs-io/staking-api-service/internal/db"
)

type V2DBClient interface {
	db.BaseDBClient
}
