package indexerdbclient

import (
	"context"
	"errors"
	"strings"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
