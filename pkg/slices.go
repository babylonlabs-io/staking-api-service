package pkg

func Map[T1, T2 any](slice []T1, f func(T1) T2) []T2 {
	if slice == nil {
		return nil
	}

	result := make([]T2, len(slice))
	for i, e := range slice {
		result[i] = f(e)
	}

	return result
}

func SliceToMap[T any, K comparable](sl []T, f func(T) K) map[K]T {
	m := make(map[K]T, len(sl))
	for _, el := range sl {
		key := f(el)
		m[key] = el
	}

	return m
}
