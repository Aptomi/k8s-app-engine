package progress

type ProgressNoop struct {
	*progressCount
}

func NewProgressNoop() *ProgressNoop {
	return &ProgressNoop{progressCount: &progressCount{}}
}

func (progressNoop *ProgressNoop) SetTotal(total int) {
	progressNoop.setTotalInternal(total)
}

func (progressNoop *ProgressNoop) Advance(stage string) {
	progressNoop.advanceInternal(stage)
}

func (progressNoop *ProgressNoop) Done() {
	progressNoop.doneInternal()
}

func (progressNoop *ProgressNoop) IsDone() bool {
	return progressNoop.isDoneInternal()
}
