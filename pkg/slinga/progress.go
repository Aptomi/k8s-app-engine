package slinga

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"time"
)

// NewProgress returns new Progress objects, to which you can add progress bars
func NewProgress() *uiprogress.Progress {
	progress := uiprogress.New()
	progress.RefreshInterval = time.Second
	progress.Start()

	return progress
}

// AddProgressBar creates a new progress bar and adds it to progress object
func AddProgressBar(progress *uiprogress.Progress, total int) *uiprogress.Bar {
	progressBar := progress.AddBar(total)
	progressBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  [%d/%d]", b.Current(), b.Total)
	})
	progressBar.AppendCompleted()
	progressBar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  Time: %s", b.TimeElapsedString())
	})

	return progressBar
}
