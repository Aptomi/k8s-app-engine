package slinga

type FakeCodeExecutor struct {
	Code *Code
}

func (executor FakeCodeExecutor) Install(key string, labels LabelSet) error {
	return nil
}

func (executor FakeCodeExecutor) Update(key string, labels LabelSet) error {
	return nil
}
func (executor FakeCodeExecutor) Destroy(key string) error {
	return nil
}
