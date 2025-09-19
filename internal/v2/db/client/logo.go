package v2dbclient

import (
	"context"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (v2 *V2Database) InsertFinalityProviderLogo(ctx context.Context, fpID string, logoURL *string) error {
	client := v2.Client.Database(v2.DbName).Collection(dbmodel.V2FinalityProvidersMetadataCollection)

	doc := v2dbmodel.FinalityProviderLogo{
		Id:        fpID,
		URL:       logoURL,
		CreatedAt: time.Now(),
	}
	_, err := client.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return &db.DuplicateKeyError{
			Key:     fpID,
			Message: err.Error(),
		}
	}

	return err
}

func (v2 *V2Database) GetFinalityProviderLogos(ctx context.Context) ([]v2dbmodel.FinalityProviderLogo, error) {
	client := v2.Client.Database(v2.DbName).Collection(dbmodel.V2FinalityProvidersMetadataCollection)

	return pkg.FetchAll[v2dbmodel.FinalityProviderLogo](ctx, client, bson.M{})
}

func (v2 *V2Database) GetFinalityProviderLogosByID(ctx context.Context, ids []string) ([]v2dbmodel.FinalityProviderLogo, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	client := v2.Client.Database(v2.DbName).Collection(dbmodel.V2FinalityProvidersMetadataCollection)
	filter := bson.M{"_id": bson.M{"$in": ids}}

	return pkg.FetchAll[v2dbmodel.FinalityProviderLogo](ctx, client, filter)
}
