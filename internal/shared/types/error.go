package types

import (
	"errors"
	"net/http"
)

type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

const (
	// 5XX
	InternalServiceError ErrorCode = "INTERNAL_SERVICE_ERROR"
	ValidationError      ErrorCode = "VALIDATION_ERROR"
	NotFound             ErrorCode = "NOT_FOUND"
	BadRequest           ErrorCode = "BAD_REQUEST"
	Forbidden            ErrorCode = "FORBIDDEN"
	UnprocessableEntity  ErrorCode = "UNPROCESSABLE_ENTITY"
	RequestTimeout       ErrorCode = "REQUEST_TIMEOUT"
)

// Error represents an error with an HTTP status code and an application-specific error code.
type Error struct {
	Err        error
	StatusCode int
	ErrorCode  ErrorCode
}

const UninitializedStatusCode = 0

func (e *Error) Error() string {
	return e.Err.Error()
}

// NewError creates a new Error with the provided status code, error code, and underlying error.
// If the status code is not provided (0), it defaults to http.StatusInternalServerError(500).
// If the error code is empty, it defaults to INTERNAL_SERVICE_ERROR.
func NewError(statusCode int, errorCode ErrorCode, err error) *Error {
	if statusCode == UninitializedStatusCode {
		statusCode = http.StatusInternalServerError
	}
	if errorCode == "" {
		errorCode = InternalServiceError
	}
	return &Error{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Err:        err,
	}
}

func NewErrorWithMsg(statusCode int, errorCode ErrorCode, msg string) *Error {
	return NewError(statusCode, errorCode, errors.New(msg))
}

func NewInternalServiceError(err error) *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  InternalServiceError,
		Err:        err,
	}
}
