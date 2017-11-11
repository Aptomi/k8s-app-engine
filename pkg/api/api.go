package api

import (
	"github.com/Aptomi/aptomi/pkg/api/codec"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/julienschmidt/httprouter"
)

type coreAPI struct {
	contentType  *codec.ContentTypeHandler
	store        store.Core
	externalData *external.Data
}

// Serve initializes everything needed by REST API and registers all API endpoints in the provided http router
func Serve(router *httprouter.Router, store store.Core, externalData *external.Data) {
	contentTypeHandler := codec.NewContentTypeHandler(runtime.NewRegistry().Append(Objects...))
	api := &coreAPI{contentTypeHandler, store, externalData}
	api.serve(router)
}

func (api *coreAPI) serve(router *httprouter.Router) {
	router.GET("/api/v1/policy", api.handlePolicyGet)
	router.GET("/api/v1/policy/gen/:gen", api.handlePolicyGet)
	router.GET("/api/v1/policy/gen/:gen/object/:ns/:kind/:name", api.handlePolicyObjectGet)
	router.POST("/api/v1/policy", api.handlePolicyUpdate)

	router.GET("/api/v1/endpoints", api.handleEndpointsGet)

	router.GET("/api/v1/revision", api.handleRevisionGet)
	router.GET("/api/v1/revision/gen/:gen", api.handleRevisionGet)
	router.GET("/api/v1/revision/policy/:policy", api.handleRevisionGetByPolicy)

	router.GET("/version", api.handleVersion)
}
