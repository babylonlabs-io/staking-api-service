package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"go.mongodb.org/mongo-driver/bson"
)

// GetFinalityProviders retrieves finality providers filtered by state
func (indexerdbclient *IndexerDatabase) GetFinalityProviders(
	ctx context.Context,
	bsnID *string,
) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	client := indexerdbclient.Client.Database(
		indexerdbclient.DbName,
	).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := bson.M{}
	if bsnID != nil && *bsnID == "all" {
		// When bsnID is "all", fetch all values without any filter
		filter = bson.M{}
	} else if bsnID != nil && *bsnID != "" {
		filter["bsn_id"] = *bsnID
	} else {
		// When bsnID is nil or empty, fetch items that don't have bsn_id field or have empty string
		// TODO: temporary solution until figure out the bsn_id for BABY chain
		filter = bson.M{
			"$or": []bson.M{
				{"bsn_id": bson.M{"$exists": false}},
				{"bsn_id": ""},
			},
		}
	}

	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*indexerdbmodel.IndexerFinalityProviderDetails
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
