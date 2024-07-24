package db

import (
	"context"
	"log"
	"time"

	utils "github.com/babylonchain/staking-api-service/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DefaultMaxAttempts   = 4 // max attempt INCLUDES the first execution
	DefaultInitialBackoff = 100 * time.Millisecond
	DefaultBackoffFactor = 2.0
)

type dbTransactionClient struct {
	*mongo.Client
}

type dbSessionWrapper struct {
	mongo.Session
}

func (c *dbTransactionClient) StartSession(opts...*options.SessionOptions) (DBSession, error) {
	session, err := c.Client.StartSession(opts...)
	if err!= nil {
		return nil, err
	}
	return &dbSessionWrapper{session}, nil
}


func (s *dbSessionWrapper) EndSession(ctx context.Context) {
	s.Session.EndSession(ctx)
}

func (s *dbSessionWrapper) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error), opts...*options.TransactionOptions) (interface{}, error) {
	return s.Session.WithTransaction(ctx, fn, opts...)
}

func TxWithRetries(
	ctx context.Context,
	dbTransactionClient DBTransactionClient,
	txnFunc func(sessCtx mongo.SessionContext) (interface{}, error),
) (interface{}, error) {
	maxAttempts := DefaultMaxAttempts
	initialBackoff := DefaultInitialBackoff
	backoffFactor := DefaultBackoffFactor

    var (
			result  interface{}
			err     error
			backoff = initialBackoff
    )

    for attempt := 1; attempt <= maxAttempts; attempt++ {
			session, sessionErr := dbTransactionClient.StartSession();


			if sessionErr != nil {
				return nil, sessionErr
			}


			result, err = session.WithTransaction(ctx, txnFunc)
			session.EndSession(ctx)

			if err != nil {
				if shouldRetry(err) && attempt < maxAttempts {
					log.Printf("Attempt %d failed with retryable error: %v. Retrying after %v...", attempt, err, backoff)
					utils.Sleep(backoff)
					backoff *= time.Duration(backoffFactor)
					continue
				}
    		log.Printf("Attempt %d failed with non-retryable error: %v", attempt, err)
    		return nil, err
			}
			break
    }
    return result, nil
}

// Check for network-related, timeout errors, write conflicts or transaction aborted, which are generally transient should retry. Other errors such as duplicated keys or other non-specified errors should be considered non-retryable.
func shouldRetry(err error) bool {
    if mongo.IsNetworkError(err) {
			return true
    }
    if mongo.IsTimeout(err) {
			return true
    }

		if IsWriteConflictError(err) {
			return true
		}

		if IsTransactionAbortedError(err) {
			return true
		}

    return false
}
