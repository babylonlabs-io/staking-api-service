package utils

import "time"

func ParseTimestampToIsoFormat(epochtime int64) string {
	// Convert the int64 epoch time to a time.Time object
	t := time.Unix(epochtime, 0)
	// Convert the time.Time object to a string in ISO8601 format
	return t.Format(time.RFC3339)
}

func GetTodayStartTimestampInSeconds() int64 {
	// Get the current time in UTC
	now := time.Now().UTC()

	// Create a new time representing today at 12AM UTC
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Convert the start of today to a Unix timestamp in seconds
	return startOfDay.Unix()
}
