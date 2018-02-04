package middleware

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/api/codec"
	"github.com/Aptomi/aptomi/pkg/runtime"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"runtime/debug"
)

type panicHandler struct {
	handler     http.Handler
	contentType *codec.ContentTypeHandler
}

// NewPanicHandler returns HTTP handler for Panics processing
func NewPanicHandler(handler http.Handler) http.Handler {
	contentTypeHandler := codec.NewContentTypeHandler(runtime.NewRegistry().Append(api.ServerErrorObject))
	return &panicHandler{handler, contentTypeHandler}
}

func (h *panicHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.WithField("request", request).Errorf("Error while serving request: %s", err)

			if log.GetLevel() >= log.DebugLevel {
				log.Debug(string(debug.Stack()))
			}

			serverErr := api.NewServerError(fmt.Sprintf("%s", err))

			h.contentType.WriteOneWithStatus(writer, request, serverErr, http.StatusInternalServerError)
		}
	}()

	h.handler.ServeHTTP(writer, request)
}
