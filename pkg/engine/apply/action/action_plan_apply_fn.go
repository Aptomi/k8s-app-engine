package action

import "sync"

// ApplyFunction is a function which applies an action
type ApplyFunction func(Base) error

// WrapSequential wraps apply function to be sequential
func WrapSequential(fn ApplyFunction) ApplyFunction {
	mutex := sync.Mutex{}
	return func(act Base) error {
		mutex.Lock()
		defer mutex.Unlock()
		return fn(act)
	}
}

// Noop is a function that does nothing
func Noop() ApplyFunction {
	return func(Base) error { return nil }
}
