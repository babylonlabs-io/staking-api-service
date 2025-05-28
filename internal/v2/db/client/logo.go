package v2dbclient

import (
	"context"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"go.mongodb.org/mongo-driver/bson"
)

func (v2 V2Database) InsertFinalityProviderLogo(ctx context.Context, identity, logoURL string) error {
	client := v2.Client.Database(v2.DbName).Collection(dbmodel.V2FinalityProvidersLogosCollection)

	doc := v2dbmodel.FinalityProviderLogo{
		Id:      identity,
		LogoURL: logoURL,
	}
	_, err := client.InsertOne(ctx, doc)
	return err
}

func (v2 *V2Database) GetFinalityProviderLogos(ctx context.Context) ([]v2dbmodel.FinalityProviderLogo, error) {
	client := v2.Client.Database(v2.DbName).Collection(dbmodel.V2FinalityProvidersLogosCollection)

	cursor, err := client.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logos []v2dbmodel.FinalityProviderLogo
	if err := cursor.All(ctx, &logos); err != nil {
		return nil, err
	}
	return logos, nil
}
