package slinga

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"time"
)

func NewProgress() *uiprogress.Progress {
	progress := uiprogress.New()
	progress.RefreshInterval = time.Second
	progress.Start()

	return progress
}

func AddProgressBar(progress *uiprogress.Progress, total int) *uiprogress.Bar {
	progressBar := progress.AddBar(total)
	progressBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  [%d/%d]", b.Current(), b.Total)
	})
	progressBar.AppendCompleted()
	progressBar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  Time: %s", strutil.PrettyTime(time.Since(b.TimeStarted)))
	})

	return progressBar
}
