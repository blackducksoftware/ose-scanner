package controller

import "time"

// Run a function and return the time.
func RunWithTime(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}
