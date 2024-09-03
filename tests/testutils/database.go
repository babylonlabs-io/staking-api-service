package testutils

import (
	"context"
	"log"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var setUpDbIndex = false

func DirectDbConnection(cfg *config.Config) *db.Database {
	client, err := mongo.Connect(
		context.TODO(), options.Client().ApplyURI(cfg.Db.Address),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &db.Database{
		DbName: cfg.Db.DbName,
		Client: client,
	}
}

// SetupTestDB connects to MongoDB and purges all collections.
func SetupTestDB(cfg config.Config) *db.Database {
	// Connect to MongoDB
	db := DirectDbConnection(&cfg)
	// Purge all collections in the test database
	// Setup the db index only once for all tests
	if !setUpDbIndex {
		err := model.Setup(context.Background(), &cfg)
		if err != nil {
			log.Fatal("Failed to setup database:", err)
		}
		setUpDbIndex = true
	}
	if err := PurgeAllCollections(context.TODO(), db.Client, cfg.Db.DbName); err != nil {
		log.Fatal("Failed to purge database:", err)
	}

	return db
}

// InjectDbDocument inserts a single document into the specified collection.
func InjectDbDocument[T any](
	cfg *config.Config, collectionName string, doc T,
) {
	connection := DirectDbConnection(cfg)
	defer connection.Client.Disconnect(context.Background())
	collection := connection.Client.Database(connection.DbName).
		Collection(collectionName)

	_, err := collection.InsertOne(context.Background(), doc)
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
}

// Inspect the items in the real database
func InspectDbDocuments[T any](
	cfg *config.Config, collectionName string,
) ([]T, error) {
	connection := DirectDbConnection(cfg)
	collection := connection.Client.Database(connection.DbName).
		Collection(collectionName)

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	defer connection.Client.Disconnect(context.Background())

	var results []T
	for cursor.Next(context.Background()) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// UpdateDbDocument updates a document in the specified collection based on the
// provided filter and update data.
func UpdateDbDocument(
	connection *db.Database, cfg *config.Config, collectionName string,
	filter bson.M, update bson.M,
) error {
	collection := connection.Client.Database(connection.DbName).
		Collection(collectionName)

	// Perform the update operation
	_, err := collection.UpdateOne(
		context.Background(), filter, bson.M{"$set": update},
	)
	if err != nil {
		return err
	}

	return nil
}

// PurgeAllCollections drops all collections in the specified database.
func PurgeAllCollections(ctx context.Context, client *mongo.Client, databaseName string) error {
	database := client.Database(databaseName)
	collections, err := database.ListCollectionNames(ctx, bson.D{{}})
	if err != nil {
		return err
	}

	for _, collection := range collections {
		// Use DeleteMany with an empty filter to delete all documents
		_, err := database.Collection(collection).DeleteMany(ctx, bson.D{{}})
		if err != nil {
			return err
		}
	}
	return nil
}
