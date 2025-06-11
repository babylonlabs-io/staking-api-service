package indexerdbclient

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	"go.mongodb.org/mongo-driver/mongo"
)

type IndexerDatabase struct {
	*dbclient.Database
}

func New(client *mongo.Client, cfg *config.DbConfig) (*IndexerDatabase, error) {
	return &IndexerDatabase{
		Database: &dbclient.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}

func (iDB *IndexerDatabase) collection(name string) *mongo.Collection {
	db := iDB.Database
	return db.Client.Database(db.DbName).Collection(name)
}
