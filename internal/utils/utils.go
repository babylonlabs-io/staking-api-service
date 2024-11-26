package utils

import (
	"encoding/json"

	"golang.org/x/exp/rand"
)

// Contains checks if a slice contains a specific element.
// It uses type parameters to work with any slice type.
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// DeepCopy performs a deep copy of a struct.
func DeepCopy(src, dst interface{}) error {
	// Marshal the source object to JSON.
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the destination object.
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}

	return nil
}

// Shuffle randomly reorders the elements in a slice.
// It uses the Fisher-Yates shuffle algorithm.
func Shuffle[T any](slice []*T) []*T {
	for i := len(slice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}
