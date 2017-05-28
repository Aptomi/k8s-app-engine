package slinga

// FakeCodeExecutor is a mock executor that does nothing
type FakeCodeExecutor struct {
	Code *Code
}

// Constructor for FakeCodeExecutor
func NewFakeCodeExecutor(code *Code) FakeCodeExecutor {
	return FakeCodeExecutor{Code: code}
}

// Install for FakeCodeExecutor does nothing
func (executor FakeCodeExecutor) Install(key string, codeMetadata map[string]string, codeParams interface{}) error {
	return nil
}

// Update for FakeCodeExecutor does nothing
func (executor FakeCodeExecutor) Update(key string, codeMetadata map[string]string, codeParams interface{}) error {
	return nil
}

// Destroy for FakeCodeExecutor does nothing
func (executor FakeCodeExecutor) Destroy(key string) error {
	return nil
}
