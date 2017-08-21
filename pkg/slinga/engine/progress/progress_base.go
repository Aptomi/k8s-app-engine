package progress

import "math"

type ProgressIndicator interface {
	// This method should be called to initialize progress indicator with 'total' steps
	SetTotal(total int)

	// This method should be called to advance progress indicator by 1 step, assuming we are located in a certain 'stage'
	Advance(stage string)

	// This method should be called when you are done using progress indicator (e.g. done, or error happened in the middle)
	Done()

	// This method should be called when you are done using progress indicator (e.g. done, or error happened in the middle)
	IsDone() bool

	// This method should be called to retrieve % of completion as integer. Note that you should rely on IsDone() instead of relying on 100% returned by this method
	GetCompletionPercent() int
}

type progressCount struct {
	stage    string
	current  int
	total    int
	finished bool
}

func (count *progressCount) setTotalInternal(total int) {
	count.total = total
}

func (count *progressCount) getTotalInternal() int {
	return count.total
}

func (count *progressCount) getStageInternal() string {
	return count.stage
}

func (count *progressCount) advanceInternal(stage string) {
	count.stage = stage
	count.current++
}

func (count *progressCount) doneInternal() {
	count.current = count.total
	count.finished = true
}

func (count *progressCount) isDoneInternal() bool {
	return count.finished
}

func (count *progressCount) GetCompletionPercent() int {
	if count.total <= 0 {
		return 0
	}
	result := int(math.Floor(100.0 * float64(count.current) / float64(count.total)))
	if result > 100 {
		result = 100
	}
	return result
}
