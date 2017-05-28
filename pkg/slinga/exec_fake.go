package slinga

import "time"

// FakeCodeExecutor is a mock executor that does nothing
type FakeCodeExecutor struct {
	Code  *Code
	Delay time.Duration
}

// Constructor for FakeCodeExecutor
func NewFakeCodeExecutor(code *Code) FakeCodeExecutor {
	return FakeCodeExecutor{Code: code}
}

// Constructor for FakeCodeExecutor
func NewFakeCodeExecutorWithDelay(code *Code, delay time.Duration) FakeCodeExecutor {
	return FakeCodeExecutor{Code: code, Delay: delay}
}

// Install for FakeCodeExecutor does nothing
func (executor FakeCodeExecutor) Install(key string, codeMetadata map[string]string, codeParams interface{}) error {
	time.Sleep(executor.Delay)
	return nil
}

// Update for FakeCodeExecutor does nothing
func (executor FakeCodeExecutor) Update(key string, codeMetadata map[string]string, codeParams interface{}) error {
	time.Sleep(executor.Delay)
	return nil
}

// Destroy for FakeCodeExecutor does nothing
func (executor FakeCodeExecutor) Destroy(key string) error {
	time.Sleep(executor.Delay)
	return nil
}
