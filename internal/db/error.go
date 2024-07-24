package db

import (
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

// DuplicateKeyError is an error type for duplicate key errors
type DuplicateKeyError struct {
	Key     string
	Message string
}

func (e *DuplicateKeyError) Error() string {
	return e.Message
}

func IsDuplicateKeyError(err error) bool {
	_, ok := err.(*DuplicateKeyError)
	return ok
}

// InvalidPaginationTokenError is an error type for invalid pagination token errors
type InvalidPaginationTokenError struct {
	Message string
}

func (e *InvalidPaginationTokenError) Error() string {
	return e.Message
}

func IsInvalidPaginationTokenError(err error) bool {
	_, ok := err.(*InvalidPaginationTokenError)
	return ok
}

// Not found Error
type NotFoundError struct {
	Key     string
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// Error code references: https://www.mongodb.com/docs/manual/reference/error-codes/
func IsWriteConflictError(err error) bool {
    if err == nil {
        log.Println("Error is nil, cannot be a write conflict")
        return false
    }

    var cmdErr *mongo.CommandError
    if errors.As(err, &cmdErr) {
        if cmdErr == nil {
            log.Println("Error is not a CommandError, cannot be a write conflict")
            return false
        }
        log.Println("Checking for write conflict error, code received:", cmdErr.Code)
        return cmdErr.Code == 112
    }

    log.Println("Error does not conform to CommandError")
    return false
}

func IsTransactionAbortedError(err error) bool {
    if err == nil {
        log.Println("Error is nil, cannot be a transaction aborted")
        return false
    }

    var cmdErr *mongo.CommandError
    if errors.As(err, &cmdErr) {
        if cmdErr == nil {
            log.Println("Error is not a CommandError, cannot be a transaction aborted")
            return false
        }
        log.Println("Checking for transaction aborted error, code received:", cmdErr.Code)
        return cmdErr.Code == 251
    }

    log.Println("Error does not conform to CommandError")
    return false
}