package indexerdbclient

import (
	"context"
	"errors"

	model "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"go.mongodb.org/mongo-driver/mongo"
)

const networkInfoID = "singleton"

type chainInfoDoc struct {
	ID               string `bson:"_id"`
	*model.ChainInfo `bson:",inline"`
}

func (idb *IndexerDatabase) GetChainInfo(
	ctx context.Context,
) (*model.ChainInfo, error) {
	filter := map[string]any{"_id": networkInfoID}
	res := idb.collection(model.NetworkInfoCollection).FindOne(ctx, filter)

	var doc chainInfoDoc
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

	return doc.ChainInfo, nil
}
