package reqresp

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// Response is the basic interface for all responses in API
type Response interface {
}

// ErrorObject contains object.Info for the Error type
var ErrorObject = &object.Info{
	Kind:        "error",
	Constructor: func() object.Base { return &Error{} },
}

// Error represents error that could be returned from the API
type Error struct {
	Metadata struct {
		Kind string
	}
	Error interface{}
}

// NewError returns instance of the error based on the provided error instance
func NewError(error interface{}) *Error {
	return &Error{struct{ Kind string }{Kind: ErrorObject.Kind}, error}
}

// GetNamespace returns namespace
func (Error) GetNamespace() string {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

// GetKind returns kind, but not supported by reqresp.Error
func (Error) GetKind() string {
	return ErrorObject.Kind
}

// GetName returns name, but not supported by reqresp.Error
func (Error) GetName() string {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

// GetGeneration returns generation, but not supported by reqresp.Error
func (Error) GetGeneration() object.Generation {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

// SetGeneration sets the generation, but not supported by reqresp.Error
func (Error) SetGeneration(object.Generation) {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}
