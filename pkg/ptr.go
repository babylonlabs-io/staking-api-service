package pkg

func Ptr[T any](v T) *T {
	return &v
}

func PtrIfNonZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}

	return &v
}
