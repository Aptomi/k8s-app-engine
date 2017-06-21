package time

import (
	"fmt"
	"math"
	"time"
)

// Diff is wrapper for time.Duration with custom methods
type Diff struct {
	duration time.Duration
}

// NewDiff wraps time.Duration into Diff
func NewDiff(duration time.Duration) *Diff {
	return &Diff{duration: duration}
}

// InSeconds returns time duration in seconds
func (d *Diff) InSeconds() int {
	return int(d.duration.Seconds())
}

// InMinutes returns time duration in minutes
func (d *Diff) InMinutes() int {
	return int(d.duration.Minutes())
}

// InHours returns time duration in hours
func (d *Diff) InHours() int {
	return int(d.duration.Hours())
}

// InDays returns time duration in days
func (d *Diff) InDays() int {
	return int(math.Floor(float64(d.InSeconds()) / 86400))
}

// Humanize returns duration as short human-readable string
func (d *Diff) Humanize() string {
	diffInSeconds := d.InSeconds()

	if diffInSeconds <= 45 {
		return fmt.Sprintf("%d sec", diffInSeconds)
	} else if diffInSeconds <= 90 {
		return "1 min"
	}

	diffInMinutes := d.InMinutes()

	if diffInMinutes <= 45 {
		return fmt.Sprintf("%d min", diffInMinutes)
	} else if diffInMinutes <= 90 {
		return "1 hour"
	}

	diffInHours := d.InHours()

	if diffInHours <= 22 {
		return fmt.Sprintf("%d hours", diffInHours)
	} else if diffInHours <= 36 {
		return "1 day"
	}

	diffInDays := d.InDays()

	return fmt.Sprintf("%d days", diffInDays)
}
