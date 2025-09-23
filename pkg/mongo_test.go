package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"testing"
)

type testItem struct {
	ID   string `bson:"_id"`
	Key1 string `bson:"key1"`
	Key2 uint64 `bson:"key2"`
}

func TestFetchAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("find error", func(t *mtest.T) {
		t.AddMockResponses(bson.D{{"ok", 0}})

		items, err := FetchAll[testItem](t.Context(), t.Coll, bson.M{})
		require.Error(t, err)
		assert.Nil(t, items)
	})
	mt.Run("decode error", func(t *mtest.T) {
		invalidDoc := mtest.CreateSuccessResponse(
			bson.E{Key: "_id", Value: "3"},
			bson.E{Key: "key1", Value: 111111}, // we expect string here
			bson.E{Key: "key2", Value: 222222},
		)

		const collectionName = "db.some_collection"

		firstBatch := mtest.CreateCursorResponse(1, collectionName, mtest.FirstBatch, invalidDoc)
		killCursors := mtest.CreateCursorResponse(0, collectionName, mtest.NextBatch)
		t.AddMockResponses(firstBatch, killCursors)

		items, err := FetchAll[testItem](t.Context(), t.Coll, bson.M{})
		require.Error(t, err)
		assert.Nil(t, items)
	})
	mt.Run("cursor error", func(t *mtest.T) {
		firstItem := testItem{
			ID:   "1",
			Key1: "key1",
			Key2: 111,
		}

		const collectionName = "db.some_collection"

		firstBatch := mtest.CreateCursorResponse(1, collectionName, mtest.FirstBatch, createItemResponse(firstItem))
		// last batch should contain killCursors, because we do not provide it here it will trigger an error in cursor.Err()
		t.AddMockResponses(firstBatch)

		items, err := FetchAll[testItem](t.Context(), t.Coll, bson.M{})
		require.Error(t, err)
		assert.Nil(t, items)
	})
	mt.Run("ok", func(t *mtest.T) {
		firstItem := testItem{
			ID:   "1",
			Key1: "key1",
			Key2: 111,
		}
		secondItem := testItem{
			ID:   "2",
			Key1: "key2",
			Key2: 222,
		}

		const collectionName = "db.some_collection"

		firstBatch := mtest.CreateCursorResponse(1, collectionName, mtest.FirstBatch, createItemResponse(firstItem))
		secondBatch := mtest.CreateCursorResponse(2, collectionName, mtest.NextBatch, createItemResponse(secondItem))
		killCursors := mtest.CreateCursorResponse(0, collectionName, mtest.NextBatch)
		t.AddMockResponses(firstBatch, secondBatch, killCursors)

		items, err := FetchAll[testItem](t.Context(), t.Coll, bson.M{})
		require.NoError(t, err)
		assert.Equal(t, []testItem{firstItem, secondItem}, items)
	})
}

func createItemResponse(item testItem) bson.D {
	return mtest.CreateSuccessResponse(
		bson.E{Key: "_id", Value: item.ID},
		bson.E{Key: "key1", Value: item.Key1},
		bson.E{Key: "key2", Value: item.Key2},
	)
}
