package indexerdbclient

import (
	"context"
	"errors"

	model "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"go.mongodb.org/mongo-driver/mongo"
)

const networkInfoID = "singleton"

type networkInfoDoc struct {
	ID                 string `bson:"_id"`
	*model.NetworkInfo `bson:",inline"`
}

func (idb *IndexerDatabase) GetNetworkInfo(ctx context.Context) (*model.NetworkInfo, error) {
	filter := map[string]any{"_id": networkInfoID}
	res := idb.collection(model.NetworkInfoCollection).FindOne(ctx, filter)

	var doc networkInfoDoc
	err := res.Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &db.NotFoundError{
				Key:     networkInfoID,
				Message: "network info not found",
			}
		}
		return nil, err
	}

	return doc.NetworkInfo, nil
}
