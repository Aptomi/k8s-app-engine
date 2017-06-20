package time

import (
	"fmt"
	"math"
	"time"
)

type Diff struct {
	duration time.Duration
}

func NewDiff(duration time.Duration) *Diff {
	return &Diff{duration: duration}
}

func (d *Diff) InSeconds() int {
	return int(d.duration.Seconds())
}

func (d *Diff) InMinutes() int {
	return int(d.duration.Minutes())
}

func (d *Diff) InHours() int {
	return int(d.duration.Hours())
}

func (d *Diff) InDays() int {
	return int(math.Floor(float64(d.InSeconds()) / 86400))
}

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
