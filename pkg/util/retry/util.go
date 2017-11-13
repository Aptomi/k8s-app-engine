package retry

import (
	"time"
)

// Func is the function to retry returning true if it's successfully completed
type Func = func() bool

// Do retries provided function "attempts" times with provided interval and returning true if it's successfully completed
func Do(attempts int, interval time.Duration, f Func) bool {
	for attempt := 0; ; attempt++ {
		if f() {
			return true
		}

		if attempt > attempts {
			break
		}

		time.Sleep(interval)
	}

	return false
}
