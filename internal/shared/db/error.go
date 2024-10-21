package db

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
