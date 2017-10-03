package progress

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"io"
	"os"
	"time"
)

type Console struct {
	*progressCount
	progress    *uiprogress.Progress
	progressBar *uiprogress.Bar
	out         io.Writer
}

func NewConsole() *Console {
	progress := uiprogress.New()
	progress.RefreshInterval = time.Second
	progress.Start()
	return &Console{
		progressCount: &progressCount{},
		progress:      progress,
		out:           os.Stdout,
	}
}

func (progressConsole *Console) createProgressBar() {
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

func (progressConsole *Console) SetOut(out io.Writer) {
	progressConsole.out = out
}

func (progressConsole *Console) SetTotal(total int) {
	progressConsole.setTotalInternal(total + 1)
	progressConsole.createProgressBar()
	progressConsole.Advance("Init")
}

func (progressConsole *Console) Advance(stage string) {
	progressConsole.advanceInternal(stage)
	progressConsole.progressBar.Incr()
}

func (progressConsole *Console) Done() {
	progressConsole.doneInternal()
	progressConsole.progress.Stop()
}

func (progressConsole *Console) IsDone() bool {
	return progressConsole.isDoneInternal()
}
