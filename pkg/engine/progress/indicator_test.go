package progress

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeProgressIndicators() map[string]Indicator {
	noop := NewNoop()
	console := NewConsole("Applying /dev/null")
	console.SetOut(new(bytes.Buffer))
	return map[string]Indicator{"noop": noop, "console": console}
}

func TestProgressNoop(t *testing.T) {
	for key, progress := range makeProgressIndicators() {
		// check initial state
		assert.False(t, progress.IsDone(), "[%s] Progress indicator should not be finished yet", key)
		assert.Equal(t, 0, progress.GetCompletionPercent(), "[%s] Progress indicator should be at 0%", key)

		// set total number of steps
		progress.SetTotal(200)
		assert.False(t, progress.IsDone(), "[%s] Progress indicator should not be finished yet", key)

		// advance
		for i := 0; i < 200; i++ {
			expectedPercent := int(math.Floor(float64(i) / 2.0))
			assert.Equal(t, expectedPercent, progress.GetCompletionPercent(), "[%s] Progress indicator should be at %d percent", key, expectedPercent)
			progress.Advance()
		}

		// check completion
		assert.Equal(t, 100, progress.GetCompletionPercent(), "[%s] Progress indicator should be at 100%", key)
		assert.False(t, progress.IsDone(), "[%s] Progress indicator should not be finished yet", key)
		progress.Done()
		assert.True(t, progress.IsDone(), "[%s] Progress indicator should be finished", key)
	}
}

func TestProgressNoopOverflow(t *testing.T) {
	for key, progress := range makeProgressIndicators() {
		progress.SetTotal(10)
		for i := 0; i < 200; i++ {
			progress.Advance()
		}
		assert.Equal(t, 100, progress.GetCompletionPercent(), "[%s] Progress indicator should be at 100%", key)
		assert.False(t, progress.IsDone(), "[%s] Progress indicator should not be finished until marked as such", key)
		progress.Done()
	}
}

func TestProgressNoopZeroLen(t *testing.T) {
	for key, progress := range makeProgressIndicators() {
		progress.SetTotal(0)
		assert.False(t, progress.IsDone(), "[%s] Progress indicator should not be finished yet", key)
		progress.Done()
		assert.True(t, progress.IsDone(), "[%s] Progress indicator should be finished", key)
	}
}
