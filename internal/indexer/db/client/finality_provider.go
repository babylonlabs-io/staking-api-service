package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"go.mongodb.org/mongo-driver/bson"
)

// GetFinalityProviderByPk retrieves a single finality provider by their primary key
func (indexerdbclient *IndexerDatabase) GetFinalityProviderByPk(
	ctx context.Context,
	fpPk string,
) (*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	client := indexerdbclient.Client.Database(
		indexerdbclient.DbName,
	).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	var result indexerdbmodel.IndexerFinalityProviderDetails
	err := client.FindOne(ctx, bson.M{"_id": fpPk}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFinalityProviders retrieves finality providers filtered by state
func (indexerdbclient *IndexerDatabase) GetFinalityProviders(
	ctx context.Context,
) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	client := indexerdbclient.Client.Database(
		indexerdbclient.DbName,
	).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	cursor, err := client.Find(ctx, bson.M{})
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
