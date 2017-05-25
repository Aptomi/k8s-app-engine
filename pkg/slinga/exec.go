package slinga

import (
	"errors"
)

// CodeExecutor is an interface that allows to create different executors for component allocation (e.g. helm, kube.libsonnet, etc)
type CodeExecutor interface {
	Install(key string, codeMetadata map[string]string, codeParams interface{}) error
	Update(key string, labels LabelSet) error
	Destroy(key string) error
}

// GetCodeExecutor returns an executor based on code.Type
func (code *Code) GetCodeExecutor() (CodeExecutor, error) {
	switch code.Type {
	case "aptomi/code/kubernetes-helm", "kubernetes-helm":
		return HelmCodeExecutor{code}, nil
	case "aptomi/code/unittests", "unittests":
		return FakeCodeExecutor{code}, nil
	default:
		return nil, errors.New("CodeExecutor not found: " + code.Type)
	}
}

