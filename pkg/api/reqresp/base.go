package reqresp

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

type Response interface {
}

var ErrorObject = &object.Info{
	Kind:        "error",
	Constructor: func() object.Base { return &Error{} },
}

type Error struct {
	Metadata struct {
		Kind string
	}
	Error interface{}
}

func NewError(error interface{}) *Error {
	return &Error{struct{ Kind string }{Kind: ""}, error}
}

func (Error) GetNamespace() string {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

func (Error) GetKind() string {
	return ErrorObject.Kind
}

func (Error) GetName() string {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

func (Error) GetGeneration() object.Generation {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

func (Error) SetGeneration(object.Generation) {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}
