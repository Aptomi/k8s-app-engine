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
	// authenticate user
	router.POST("/api/v1/user/authenticate", api.authenticateUser)

	// get all users and their roles
	router.GET("/api/v1/user/roles", api.handleUserRoles)

	// retrieve policy (latest + by a given generation)
	router.GET("/api/v1/policy", api.handlePolicyGet)
	router.GET("/api/v1/policy/gen/:gen", api.handlePolicyGet)

	// retrieve specific object from the policy
	router.GET("/api/v1/policy/gen/:gen/object/:ns/:kind/:name", api.handlePolicyObjectGet)

	// update policy
	router.POST("/api/v1/policy", api.handlePolicyUpdate)

	// policy diagrams
	router.GET("/api/v1/policy/diagram", api.handlePolicyDiagram)
	router.GET("/api/v1/policy/diagram/gen/:gen", api.handlePolicyDiagram)
	router.GET("/api/v1/policy/diagram/compare/gen/:gen/genBase/:genBase", api.handlePolicyDiagramCompare)

	// retrieve endpoints
	router.GET("/api/v1/endpoints", api.handleEndpointsGet)

	// retrieve revision (latest + by a given generation)
	router.GET("/api/v1/revision", api.handleRevisionGet)
	router.GET("/api/v1/revision/gen/:gen", api.handleRevisionGet)

	// retrieve revision(s) (for a given policy)
	router.GET("/api/v1/revision/policy/:policy", api.handleRevisionGetByPolicy)
	router.GET("/api/v1/revisions/policy/:policy", api.handleRevisionsGetByPolicy)

	router.DELETE("/api/v1/actualstate", api.handleActualStateReset)

	// return aptomi version
	router.GET("/version", api.handleVersion)
}
