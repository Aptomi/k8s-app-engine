package retry

import (
	"fmt"
	"time"
)

// Func is the function to retry returning true if it's successfully completed
type Func = func() bool

// Do retries provided function until maxTime is reached, using provided interval as a delay and returning true if it's successfully completed
func Do(maxTime time.Duration, interval time.Duration, f Func) bool {
	if interval < 100*time.Millisecond {
		panic(fmt.Sprintf("retry.Do used with interval less then 1/10 second, it seems dangerous: %s", interval))
	}

	start := time.Now()
	for time.Since(start) < maxTime {
		if f() {
			return true
		}

		// sleep
		time.Sleep(interval)
	}

	return false
}

// Do2 retries provided until maxTime is reached, doubling the delay interval until it reaches the maxInterval and returning true if it's successfully completed
func Do2(maxTime time.Duration, maxInterval time.Duration, f Func) bool {
	if maxInterval < 100*time.Millisecond {
		panic(fmt.Sprintf("retry.Do2 used with maxInterval less then 1/10 second, it seems dangerous: %s", maxInterval))
	}

	start := time.Now()
	interval := 100 * time.Millisecond
	for time.Since(start) < maxTime {
		if f() {
			return true
		}

		// sleep
		time.Sleep(interval)

		// double the interval until it reaches maxInterval
		interval *= 2
		if interval > maxInterval {
			interval = maxInterval
		}
	}

	return false
}
