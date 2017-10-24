package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"github.com/Aptomi/aptomi/pkg/object"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"runtime/debug"
)

// newPanicHandler returns HTTP handler for Panics processing
func newPanicHandler(handler http.Handler) http.Handler {
	return &panicHandler{handler}
}

type panicHandler struct {
	handler http.Handler
}

func (h *panicHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			log.WithField("request", req).Errorf("Error while serving request: %s", err)

			if log.GetLevel() >= log.DebugLevel {
				log.Debug(string(debug.Stack()))
			}

			data := yaml.SerializeObject(newServerError(err))
			_, wErr := fmt.Fprint(w, data)
			if wErr != nil {
				log.Errorf("Error while writing error to response: %s", err)
			}
		}
	}()

	h.handler.ServeHTTP(w, req)
}

// serverErrorObject contains object.Info for the Error type
var serverErrorObject = &object.Info{
	Kind:        "error",
	Constructor: func() object.Base { return &serverError{} },
}

// Error represents error that could be returned from the API
type serverError struct {
	Metadata struct {
		Kind string
	}
	Error interface{}
}

// newServerError returns instance of the error based on the provided error instance
func newServerError(error interface{}) *serverError {
	return &serverError{struct{ Kind string }{Kind: serverErrorObject.Kind}, error}
}

// GetNamespace returns namespace
func (serverError) GetNamespace() string {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

// GetKind returns kind, but not supported by reqresp.Error
func (serverError) GetKind() string {
	return serverErrorObject.Kind
}

// GetName returns name, but not supported by reqresp.Error
func (serverError) GetName() string {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

// GetGeneration returns generation, but not supported by reqresp.Error
func (serverError) GetGeneration() object.Generation {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}

// SetGeneration sets the generation, but not supported by reqresp.Error
func (serverError) SetGeneration(object.Generation) {
	panic("reqresp.Error only mimics object.Base, so, only GetKind() could be called")
}
