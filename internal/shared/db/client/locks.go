package dbclient

import (
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Database) LocksCollection() *mongo.Collection {
	return db.Client.Database(db.DbName).Collection(dbmodel.Locks)
}
