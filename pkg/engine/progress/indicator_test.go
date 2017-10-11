package progress

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func makeProgressIndicators() []Indicator {
	p1 := NewNoop()
	p2 := NewConsole()
	p2.SetOut(new(bytes.Buffer))
	return []Indicator{p1, p2}
}

func TestProgressNoop(t *testing.T) {
	for _, progress := range makeProgressIndicators() {
		// check initial state
		assert.False(t, progress.IsDone(), "Progress indicator should not be finished yet")
		assert.Equal(t, 0, progress.GetCompletionPercent(), "Progress indicator should be at 0%")

		// set total number of steps
		progress.SetTotal(200)
		assert.False(t, progress.IsDone(), "Progress indicator should not be finished yet")

		// advance
		for i := 0; i < 200; i++ {
			expectedPercent := int(math.Floor(float64(i) / 2.0))
			assert.Equal(t, expectedPercent, progress.GetCompletionPercent(), "Progress indicator should be at 25%")
			progress.Advance("Main")
		}

		// check completion
		assert.Equal(t, 100, progress.GetCompletionPercent(), "Progress indicator should be at 100%")
		assert.False(t, progress.IsDone(), "Progress indicator should not be finished yet")
		progress.Done()
		assert.True(t, progress.IsDone(), "Progress indicator should be finished")
	}
}

func TestProgressNoopOverflow(t *testing.T) {
	for _, progress := range makeProgressIndicators() {
		progress.SetTotal(10)
		for i := 0; i < 200; i++ {
			progress.Advance("Main")
		}
		assert.Equal(t, 100, progress.GetCompletionPercent(), "Progress indicator should be at 100%")
		assert.False(t, progress.IsDone(), "Progress indicator should not be finished until marked as such")
		progress.Done()
	}
}

func TestProgressNoopZeroLen(t *testing.T) {
	for _, progress := range makeProgressIndicators() {
		progress.SetTotal(0)
		assert.False(t, progress.IsDone(), "Progress indicator should not be finished yet")
		progress.Done()
		assert.True(t, progress.IsDone(), "Progress indicator should be finished")
	}
}
