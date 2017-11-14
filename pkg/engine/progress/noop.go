package progress

// Noop is a mock progress indicator which prints nothing
type Noop struct {
	*progressCount
}

// NewNoop creates a new noop progress indicator
func NewNoop() *Noop {
	return &Noop{progressCount: &progressCount{}}
}

// SetTotal sets the total number of steps in a progress indicator
func (progressNoop *Noop) SetTotal(total int) {
	progressNoop.setTotalInternal(total)
}

// Advance advances progress indicator by one step
func (progressNoop *Noop) Advance() {
	progressNoop.advanceInternal()
}

// Done should be called once done working with progress indicator
func (progressNoop *Noop) Done(success bool) {
	progressNoop.doneInternal()
}

// IsDone returns if progress indicator was already marked as Done()
func (progressNoop *Noop) IsDone() bool {
	return progressNoop.isDoneInternal()
}
