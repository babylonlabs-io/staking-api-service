package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbResultMap[T any] struct {
	Data            []T    `json:"data"`
	PaginationToken string `json:"paginationToken"`
}

/*
Builds the result map with a pagination token.
If the result length exceeds the maximum limit, it returns the map with a token.
Otherwise, it returns the map with an empty token. Note that the pagination
limit is the maximum number of results to return.
For example, if the limit is 10, it fetches 11 but returns only 10.
The last result is used to generate the pagination token.
*/
func toResultMapWithPaginationToken[T any](paginationLimit int64, result []T, paginationKeyBuilder func(T) (string, error)) (*DbResultMap[T], error) {
	if len(result) > int(paginationLimit) {
		result = result[:paginationLimit]
		paginationToken, err := paginationKeyBuilder(result[len(result)-1])
		if err != nil {
			return nil, err
		}
		return &DbResultMap[T]{
			Data:            result,
			PaginationToken: paginationToken,
		}, nil
	}

	return &DbResultMap[T]{
		Data:            result,
		PaginationToken: "",
	}, nil
}

// Finds documents in the collection with pagination in returned results.
func FindWithPagination[T any](
	ctx context.Context, client *mongo.Collection, filter bson.M,
	options *options.FindOptions, limit int64,
	paginationKeyBuilder func(T) (string, error),
) (*DbResultMap[T], error) {
	// Always fetch one more than the limit to check if there are more results
	// this is used to generate the pagination token
	options.SetLimit(limit + 1)

	cursor, err := client.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []T
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return toResultMapWithPaginationToken(limit, result, paginationKeyBuilder)
}
