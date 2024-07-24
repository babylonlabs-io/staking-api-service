package tests

import (
	"context"
	"testing"
	"time"
    utils "github.com/babylonchain/staking-api-service/internal/utils"
	db "github.com/babylonchain/staking-api-service/internal/db"
	testmock "github.com/babylonchain/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func writeConflictError() *mongo.CommandError {
    return &mongo.CommandError{
        Code:    112,
        Message: "write conflict",
        Name:    "WriteConflict",
    }
}

func TestTxWithRetries_ExponentialBackoff(t *testing.T) {
    mockDBClient := testmock.NewDBTransactionClient(t)

    mockSession := testmock.NewDBSession(t)

    // Simulate a txnFunc that will fail on the first two attempts
    txnFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
        return nil, writeConflictError()
    }

    // Define the session handling and withTransaction behaviors
    mockDBClient.On("StartSession").Return(mockSession, nil)
    mockSession.On("WithTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil, writeConflictError()).Twice()
    mockSession.On("WithTransaction", mock.Anything, mock.Anything, mock.Anything).Return("success", nil).Once()
    mockSession.On("EndSession", mock.Anything).Return()

    sleepDurations := []time.Duration{}
    utils.SetSleepFunc(func(d time.Duration) {
        sleepDurations = append(sleepDurations, d)
    })
    defer utils.ResetSleepFunc()

    // Execute the function that includes the retry logic
    result, err := db.TxWithRetries(context.Background(), mockDBClient, txnFunc)

    require.NoError(t, err)
    require.Equal(t, "success", result)

    mockSession.AssertCalled(t, "EndSession", mock.Anything)

    expectedBackoffDurations := []time.Duration{
        100 * time.Millisecond,
        200 * time.Millisecond,
    }

    require.Equal(t, expectedBackoffDurations, sleepDurations)
}

func TestTxWithRetries_MaxRetries(t *testing.T) {
    mockDBClient := testmock.NewDBTransactionClient(t)
    mockSession := testmock.NewDBSession(t)

    txnFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
        return nil, writeConflictError()
    }

    mockDBClient.On("StartSession").Return(mockSession, nil)
    mockSession.On("WithTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil, writeConflictError()).Times(4)  // Assuming max attempts is 4
    mockSession.On("EndSession", mock.Anything).Return()

    sleepDurations := []time.Duration{}
    utils.SetSleepFunc(func(d time.Duration) {
        sleepDurations = append(sleepDurations, d)
    })
    defer utils.ResetSleepFunc()

    // Execute the function
    result, err := db.TxWithRetries(context.Background(), mockDBClient, txnFunc)

    require.Error(t, err)
    require.Nil(t, result)
    require.Len(t, sleepDurations, 3)

    mockSession.AssertExpectations(t)
}

func TestTxWithRetries_NonRetryableError(t *testing.T) {
    mockDBClient := testmock.NewDBTransactionClient(t)
    mockSession := testmock.NewDBSession(t)

    nonRetryableError := &mongo.CommandError{
        Code:    403,  // Example error code that is not retryable
        Message: "Forbidden",
        Name:    "NonRetryableError",
    }

    txnFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
        return nil, nonRetryableError
    }

    // Define session handling for non-retryable error
    mockDBClient.On("StartSession").Return(mockSession, nil)
    mockSession.On("WithTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil, nonRetryableError).Once()
    mockSession.On("EndSession", mock.Anything).Return()

    sleepDurations := []time.Duration{}
    utils.SetSleepFunc(func(d time.Duration) {
        sleepDurations = append(sleepDurations, d)
    })
    defer utils.ResetSleepFunc()

    // Execute the function
    result, err := db.TxWithRetries(context.Background(), mockDBClient, txnFunc)

    require.Error(t, err)
    require.Nil(t, result)
    require.Len(t, sleepDurations, 0)  // Ensure no retries occurred

    require.IsType(t, nonRetryableError, err)
    mockSession.AssertExpectations(t)
}