package progress

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"io"
	"os"
	"time"
)

// Console is a console-based progress indicator
type Console struct {
	*progressCount
	progress    *uiprogress.Progress
	progressBar *uiprogress.Bar
	out         io.Writer
}

// NewConsole creates a new console-based progress indicator
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
		return fmt.Sprintf("  [%d/%d]", b.Current(), b.Total)
	})
	progressConsole.progressBar.AppendCompleted()
	progressConsole.progressBar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("  Time: %s", b.TimeElapsedString())
	})
}

// SetOut set output writer for writing progress information to
func (progressConsole *Console) SetOut(out io.Writer) {
	progressConsole.out = out
}

// SetTotal sets the total number of steps in a progress indicator
func (progressConsole *Console) SetTotal(total int) {
	progressConsole.setTotalInternal(total + 1)
	progressConsole.createProgressBar()
	progressConsole.Advance()
}

// Advance advances progress indicator by one step
func (progressConsole *Console) Advance() {
	progressConsole.advanceInternal()
	progressConsole.progressBar.Incr()
}

// Done should be called once done working with progress indicator
func (progressConsole *Console) Done(success bool) {
	progressConsole.doneInternal()
	progressConsole.progress.Stop()
}

// IsDone returns if progress indicator was already marked as Done()
func (progressConsole *Console) IsDone() bool {
	return progressConsole.isDoneInternal()
}
