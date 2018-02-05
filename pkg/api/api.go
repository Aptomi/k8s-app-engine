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
	secret       string
}

// Serve initializes everything needed by REST API and registers all API endpoints in the provided http router
func Serve(router *httprouter.Router, store store.Core, externalData *external.Data, secret string) {
	contentTypeHandler := codec.NewContentTypeHandler(runtime.NewRegistry().Append(Objects...))
	api := &coreAPI{contentTypeHandler, store, externalData, secret}
	api.serve(router)
}

func (api *coreAPI) serve(router *httprouter.Router) {
	auth := api.auth

	// authenticate user
	router.POST("/api/v1/user/login", api.handleLogin)

	// get all users and their roles
	router.GET("/api/v1/user/roles", auth(api.handleUserRoles))

	// retrieve policy (latest + by a given generation)
	router.GET("/api/v1/policy", auth(api.handlePolicyGet))
	router.GET("/api/v1/policy/gen/:gen", auth(api.handlePolicyGet))

	// retrieve specific object from the policy
	router.GET("/api/v1/policy/gen/:gen/object/:ns/:kind/:name", auth(api.handlePolicyObjectGet))

	// update policy
	router.POST("/api/v1/policy", auth(api.handlePolicyUpdate))
	router.DELETE("/api/v1/policy", auth(api.handlePolicyDelete))

	// policy diagrams
	router.GET("/api/v1/policy/diagram/mode/:mode", auth(api.handlePolicyDiagram))
	router.GET("/api/v1/policy/diagram/mode/:mode/gen/:gen", auth(api.handlePolicyDiagram))
	router.GET("/api/v1/policy/diagram/compare/mode/:mode/gen/:gen/genBase/:genBase", auth(api.handlePolicyDiagramCompare))

	// retrieve dependency along with its status
	router.GET("/api/v1/policy/dependency/:ns/:name/status", auth(api.handleDependencyStatusGet))

	// retrieve endpoints (all + by dependency)
	router.GET("/api/v1/endpoints", api.handleEndpointsGet)
	router.GET("/api/v1/endpoints/dependency/:ns/:name", auth(api.handleEndpointsGet))

	// retrieve revision (latest + by a given generation)
	router.GET("/api/v1/revision", auth(api.handleRevisionGet))
	router.GET("/api/v1/revision/gen/:gen", auth(api.handleRevisionGet))

	// retrieve revision(s) (for a given policy)
	router.GET("/api/v1/revision/policy/:policy", auth(api.handleRevisionGetByPolicy))
	router.GET("/api/v1/revisions/policy/:policy", auth(api.handleRevisionsGetByPolicy))

	router.DELETE("/api/v1/actualstate", auth(api.handleActualStateReset))

	// return aptomi version
	router.GET("/version", api.handleVersion)
	router.GET("/api/v1/version", api.handleVersion)
}
