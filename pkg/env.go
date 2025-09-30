package pkg

import "os"

func Getenv(key, defaultValue string) string {
	value, found := os.LookupEnv(key)
	if !found {
		value = defaultValue
	}

	return value
}
