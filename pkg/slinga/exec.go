package slinga

import (
	"errors"
	"time"
)

// CodeExecutor is an interface that allows to create different executors for component allocation (e.g. helm, kube.libsonnet, etc)
type CodeExecutor interface {
	Install(key string, codeMetadata map[string]string, codeParams interface{}) error
	Update(key string, codeMetadata map[string]string, codeParams interface{}) error
	Destroy(key string) error
}

// GetCodeExecutor returns an executor based on code.Type
func (code *Code) GetCodeExecutor() (CodeExecutor, error) {
	switch code.Type {
	case "aptomi/code/kubernetes-helm", "kubernetes-helm":
		if kubeClient, ok := code.cluster.Client().(*KubeClient); ok {
			return NewHelmCodeExecutor(code, kubeClient.tillerHost), nil
		}
		return nil, errors.New("Helm executor works only with K8s cluster")
	case "aptomi/code/unittests", "unittests":
		return NewFakeCodeExecutor(code), nil
	case "aptomi/code/withdelay", "delay":
		return NewFakeCodeExecutorWithDelay(code, time.Second), nil
	default:
		return nil, errors.New("CodeExecutor not found: " + code.Type)
	}
}
