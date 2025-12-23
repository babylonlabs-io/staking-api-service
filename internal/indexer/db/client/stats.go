package indexerdbclient

import (
	"context"
	"errors"
	"strings"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetOverallStats fetches the overall stats from the indexer's stats collection
func (iDB *IndexerDatabase) GetOverallStats(ctx context.Context) (*indexerdbmodel.IndexerStatsDocument, error) {
	collection := iDB.collection(dbmodel.IndexerStatsCollection)
	filter := bson.M{"_id": "overall_stats"}

	var stats indexerdbmodel.IndexerStatsDocument
	err := collection.FindOne(ctx, filter).Decode(&stats)
	if err != nil {
		// If no stats found, return zeros
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &indexerdbmodel.IndexerStatsDocument{
				Id:                "overall_stats",
				ActiveTvl:         0,
				ActiveDelegations: 0,
				LastUpdated:       0,
			}, nil
		}
		return nil, err
	}

	return &stats, nil
}

// GetFinalityProviderStats fetches finality provider stats from the indexer's collection
func (iDB *IndexerDatabase) GetFinalityProviderStats(
	ctx context.Context,
	fpPkHexes []string,
) ([]*indexerdbmodel.IndexerFinalityProviderStatsDocument, error) {
	if len(fpPkHexes) == 0 {
		return nil, nil
	}

	// Convert to lowercase for case-insensitive matching (indexer stores lowercase)
	var lowercaseFpPkHexes []string
	for _, fpPkHex := range fpPkHexes {
		lowercaseFpPkHexes = append(lowercaseFpPkHexes, strings.ToLower(fpPkHex))
	}

	collection := iDB.collection(dbmodel.IndexerFinalityProviderStatsCollection)
	filter := bson.M{"_id": bson.M{"$in": lowercaseFpPkHexes}}

	return pkg.FetchAll[*indexerdbmodel.IndexerFinalityProviderStatsDocument](ctx, collection, filter)
}

// GetFinalityProviderStatsPaginated retrieves finality provider stats sorted by active_tvl DESC
// This method is used for the V2 finality providers endpoint to enable sorting by TVL
func (iDB *IndexerDatabase) GetFinalityProviderStatsPaginated(
	ctx context.Context,
	paginationToken string,
) (*db.DbResultMap[*indexerdbmodel.IndexerFinalityProviderStatsDocument], error) {
	collection := iDB.collection(dbmodel.IndexerFinalityProviderStatsCollection)
	opts := options.Find().SetSort(bson.D{
		{Key: "active_tvl", Value: -1},
		{Key: "_id", Value: -1},
	})
	var filter bson.M

	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerFinalityProviderStatsPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter = bson.M{
			"$or": []bson.M{
				{"active_tvl": bson.M{"$lt": decodedToken.ActiveTvl}},
				{"active_tvl": decodedToken.ActiveTvl, "_id": bson.M{"$lt": strings.ToLower(decodedToken.FpBtcPkHex)}},
			},
		}
	}

	return db.FindWithPagination(
		ctx,
		collection,
		filter,
		opts,
		iDB.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildIndexerFinalityProviderStatsPaginationToken,
	)
}
