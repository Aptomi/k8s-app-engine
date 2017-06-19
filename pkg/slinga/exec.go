package slinga

import (
	"errors"
	"time"
)

// CodeExecutor is an interface that allows to create different executors for component allocation (e.g. helm, kube.libsonnet, etc)
type CodeExecutor interface {
	Install() error
	Update() error
	Endpoints() (map[string]string, error)
	Destroy() error
}

// GetCodeExecutor returns an executor based on code.Type
func (code *Code) GetCodeExecutor(key string, codeMetadata map[string]string, codeParams NestedParameterMap, clusters map[string]*Cluster) (CodeExecutor, error) {
	switch code.Type {
	case "aptomi/code/kubernetes-helm", "kubernetes-helm":
		return NewHelmCodeExecutor(code, key, codeMetadata, codeParams, clusters)
	case "aptomi/code/unittests", "unittests":
		return NewFakeCodeExecutor(code, key, codeMetadata, codeParams, clusters), nil
	case "aptomi/code/withdelay", "delay":
		return NewFakeCodeExecutorWithDelay(code, key, codeMetadata, codeParams, clusters, 100*time.Millisecond), nil
	default:
		return nil, errors.New("CodeExecutor not found: " + code.Type)
	}
}
