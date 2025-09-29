package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"go.mongodb.org/mongo-driver/bson"
)

// GetFinalityProviders retrieves finality providers filtered by state
func (indexerdbclient *IndexerDatabase) GetFinalityProviders(
	ctx context.Context,
) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	client := indexerdbclient.Client.Database(
		indexerdbclient.DbName,
	).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	// Fetch all finality providers
	filter := bson.M{}

	return pkg.FetchAll[*indexerdbmodel.IndexerFinalityProviderDetails](ctx, client, filter)
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

	return pkg.FetchAll[*indexerdbmodel.IndexerFinalityProviderDetails](ctx, client, filter)
}
