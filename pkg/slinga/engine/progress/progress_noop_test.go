package progress

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"math"
)

func TestProgressNoop(t *testing.T) {
	progress := NewProgressNoop()

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

func TestProgressNoopOverflow(t *testing.T) {
	progress := NewProgressNoop()
	progress.SetTotal(10)
	for i := 0; i < 200; i++ {
		progress.Advance("Main")
	}
	assert.Equal(t, 100, progress.GetCompletionPercent(), "Progress indicator should be at 100%")
	assert.False(t, progress.IsDone(), "Progress indicator should not be finished until marked as such")
}
