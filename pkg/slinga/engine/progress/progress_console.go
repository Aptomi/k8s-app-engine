package progress

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"io"
	"os"
	"time"
)

type ProgressConsole struct {
	*progressCount
	progress    *uiprogress.Progress
	progressBar *uiprogress.Bar
	out         io.Writer
}

func NewProgressConsole() *ProgressConsole {
	progress := uiprogress.New()
	progress.RefreshInterval = time.Second
	progress.Start()
	return &ProgressConsole{
		progressCount: &progressCount{},
		progress:      progress,
		out:           os.Stdout,
	}
}

func (progressConsole *ProgressConsole) createProgressBar() {
	progressConsole.progress.SetOut(progressConsole.out)
	if progressConsole.getTotalInternal() > 0 {
		fmt.Fprintln(progressConsole.out, "[Applying changes]")
	}
	progressConsole.progressBar = progressConsole.progress.AddBar(progressConsole.getTotalInternal())
	progressConsole.progressBar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  [%s: %d/%d]", progressConsole.getStageInternal(), b.Current(), b.Total)
	})
	progressConsole.progressBar.AppendCompleted()
	progressConsole.progressBar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  Time: %s", b.TimeElapsedString())
	})
}

func (progressConsole *ProgressConsole) SetOut(out io.Writer) {
	progressConsole.out = out
}

func (progressConsole *ProgressConsole) SetTotal(total int) {
	progressConsole.setTotalInternal(total + 1)
	progressConsole.createProgressBar()
	progressConsole.Advance("Init")
}

func (progressConsole *ProgressConsole) Advance(stage string) {
	progressConsole.advanceInternal(stage)
	progressConsole.progressBar.Incr()
}

func (progressConsole *ProgressConsole) Done() {
	progressConsole.doneInternal()
	progressConsole.progress.Stop()
}

func (progressConsole *ProgressConsole) IsDone() bool {
	return progressConsole.isDoneInternal()
}
