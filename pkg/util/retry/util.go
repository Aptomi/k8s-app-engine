package retry

import (
	"fmt"
	"time"
)

// Func is the function to retry returning true if it's successfully completed
type Func = func() bool

// Do retries provided function "attempts" times with provided interval and returning true if it's successfully completed
func Do(attempts int, interval time.Duration, f Func) bool {
	if interval < 100*time.Millisecond {
		panic(fmt.Sprintf("retry.Do used with interval less then 1/10 second, it seems dangerous: %s", interval))
	}

	for attempt := 0; attempt < attempts; attempt++ {
		if f() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}

// Do2 retries provided function "attempts" times, doubling the interval until it reaches the maxInterval and returning true if it's successfully completed
func Do2(attempts int, maxInterval time.Duration, f Func) bool {
	if maxInterval < 100*time.Millisecond {
		panic(fmt.Sprintf("retry.Do2 used with maxInterval less then 1/10 second, it seems dangerous: %s", maxInterval))
	}

	interval := 100 * time.Millisecond
	for attempt := 0; attempt < attempts; {
		if f() {
			return true
		}

		// sleep
		time.Sleep(interval)

		// double the interval until it reaches maxInterval
		if interval < maxInterval {
			interval *= 2
			if interval > maxInterval {
				interval = maxInterval
			}
		} else {
			// if it already reached maxInterval, start counting attempts
			attempt++
		}
	}

	return false
}
