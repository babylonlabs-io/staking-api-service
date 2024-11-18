package v2dbclient

import (
	"context"

	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
)

type V2DBClient interface {
	dbclient.DBClient
	GetOverallStats(ctx context.Context) (*v2dbmodel.V2OverallStatsDocument, error)
	GetStakerStats(ctx context.Context, stakerPKHex string) (*v2dbmodel.V2StakerStatsDocument, error)
}
