package progress

type Noop struct {
	*progressCount
}

func NewNoop() *Noop {
	return &Noop{progressCount: &progressCount{}}
}

func (progressNoop *Noop) SetTotal(total int) {
	progressNoop.setTotalInternal(total)
}

func (progressNoop *Noop) Advance(stage string) {
	progressNoop.advanceInternal(stage)
}

func (progressNoop *Noop) Done() {
	progressNoop.doneInternal()
}

func (progressNoop *Noop) IsDone() bool {
	return progressNoop.isDoneInternal()
}
