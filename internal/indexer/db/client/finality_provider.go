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

	// default filter to fetch all finality providers if bsnID is nil
	filter := bson.M{}
	if bsnID != nil {
		if *bsnID == "all" {
			// When bsnID is "all", fetch all values without any filter
			filter = bson.M{}
		} else {
			filter["bsn_id"] = *bsnID
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

// GetFinalityProvidersByID retrieves finality providers by their id-s
func (indexerdbclient *IndexerDatabase) GetFinalityProvidersByID(
	ctx context.Context,
	ids []string,
) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	client := indexerdbclient.collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := bson.M{"_id": bson.M{"$in": ids}}
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
