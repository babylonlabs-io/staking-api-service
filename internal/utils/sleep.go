package utils 

import (
	"sync"
	"time"
)

var (
	sleepFunc func(time.Duration)
	mu        sync.Mutex // mutex to make the setting of the sleepFunc thread-safe
)

func init() {
	ResetSleepFunc() // Initialize sleepFunc with the default sleep function
}

// Sleep calls the current sleep function.
func Sleep(d time.Duration) {
	mu.Lock()
	f := sleepFunc
	mu.Unlock()
	f(d)
}

// SetSleepFunc allows for overriding the default sleep function, primarily for testing.
func SetSleepFunc(f func(time.Duration)) {
	mu.Lock()
	sleepFunc = f
	mu.Unlock()
}

// ResetSleepFunc resets the sleep function to the default time.Sleep.
func ResetSleepFunc() {
	SetSleepFunc(time.Sleep)
}
