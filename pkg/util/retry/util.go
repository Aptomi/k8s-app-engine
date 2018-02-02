package retry

import (
	"fmt"
	"time"
)

// Func is the function to retry returning true if it's successfully completed
type Func = func() bool

// Do retries provided function "attempts" times with provided interval and returning true if it's successfully completed
func Do(attempts int, interval time.Duration, f Func) bool {
	if interval < 1*time.Second/10 {
		panic(fmt.Sprintf("retry.Do used with interval less then 1/10 second, it seems dangerous: %s", interval))
	}

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
