package lang

import "github.com/Aptomi/aptomi/pkg/runtime"

// Base interface represents unified base object that could be part of the policy
type Base interface {
	runtime.Deletable
}
