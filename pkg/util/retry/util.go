package retry

import "time"

type Func = func() bool

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
