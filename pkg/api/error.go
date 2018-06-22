package api

import "github.com/Aptomi/aptomi/pkg/runtime"

// ServerErrorObject contains Info for the Error type
var ServerErrorObject = &runtime.Info{
	Kind:        "error",
	Constructor: func() runtime.Object { return &ServerError{} },
}

// ServerError represents error that could be returned from the API
type ServerError struct {
	runtime.TypeKind `yaml:",inline"`
	Error            string
}

// NewServerError returns instance of the error based on the provided error
func NewServerError(error string) *ServerError {
	return &ServerError{TypeKind: ServerErrorObject.GetTypeKind(), Error: error}
}
