package slinga

import (
	"errors"
	. "github.com/Frostman/aptomi/pkg/slinga/language"
	. "github.com/Frostman/aptomi/pkg/slinga/util"
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
func GetCodeExecutor(code *Code, key string, codeParams NestedParameterMap, clusters map[string]*Cluster) (CodeExecutor, error) {
	switch code.Type {
	case "aptomi/code/kubernetes-helm", "kubernetes-helm":
		return NewHelmCodeExecutor(code, key, codeParams, clusters)
	case "aptomi/code/unittests", "unittests":
		return NewFakeCodeExecutor(code, key, codeParams, clusters), nil
	case "aptomi/code/withdelay", "delay":
		return NewFakeCodeExecutorWithDelay(code, key, codeParams, clusters, 100*time.Millisecond), nil
	default:
		return nil, errors.New("CodeExecutor not found: " + code.Type)
	}
}
