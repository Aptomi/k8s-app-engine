package api

import (
	"sync"

	"github.com/Aptomi/aptomi/pkg/api/codec"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/registry"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type coreAPI struct {
	contentType                  *codec.ContentTypeHandler
	registry                     registry.Interface
	externalData                 *external.Data
	pluginRegistryFactory        plugin.RegistryFactory
	secret                       string
	logLevel                     logrus.Level
	runDesiredStateEnforcement   chan bool
	policyAndRevisionUpdateMutex sync.Mutex
}

// Serve initializes everything needed by REST API and registers all API endpoints in the provided http router
func Serve(router *httprouter.Router, registry registry.Interface, externalData *external.Data, pluginRegistryFactory plugin.RegistryFactory, secret string, logLevel logrus.Level, runDesiredStateEnforcement chan bool) {
	contentTypeHandler := codec.NewContentTypeHandler(runtime.NewTypes().Append(Types...))
	api := &coreAPI{
		contentType:                contentTypeHandler,
		registry:                   registry,
		externalData:               externalData,
		pluginRegistryFactory:      pluginRegistryFactory,
		secret:                     secret,
		logLevel:                   logLevel,
		runDesiredStateEnforcement: runDesiredStateEnforcement,
	}
	api.serve(router)
}

func (api *coreAPI) serve(router *httprouter.Router) {
	auth := api.auth

	// todo consider moving to a separate port for security (should be nothing sensetive?)
	// prometheus metrics handler
	router.Handler("GET", "/metrics", promhttp.Handler())

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
	router.POST("/api/v1/policy/noop/:noop/loglevel/:loglevel", auth(api.handlePolicyUpdate))
	router.DELETE("/api/v1/policy", auth(api.handlePolicyDelete))
	router.DELETE("/api/v1/policy/noop/:noop/loglevel/:loglevel", auth(api.handlePolicyDelete))

	// policy & object diagrams
	router.GET("/api/v1/policy/diagram/object/:ns/:kind/:name", auth(api.handleObjectDiagram))
	router.GET("/api/v1/policy/diagram/mode/:mode", auth(api.handlePolicyDiagram))
	router.GET("/api/v1/policy/diagram/mode/:mode/gen/:gen", auth(api.handlePolicyDiagram))
	router.GET("/api/v1/policy/diagram/compare/mode/:mode/gen/:gen/genBase/:genBase", auth(api.handlePolicyDiagramCompare))

	// retrieve claim along with its status
	router.GET("/api/v1/policy/claim/status/:queryFlag/:idList", auth(api.handleClaimStatusGet))
	router.GET("/api/v1/policy/claim/resources/:ns/:name", auth(api.handleClaimResourcesGet))

	// retrieve revision (latest + by a given generation)
	router.GET("/api/v1/revision", auth(api.handleRevisionGet))
	router.GET("/api/v1/revision/gen/:gen", auth(api.handleRevisionGet))

	// retrieve revision(s) (for a given policy)
	router.GET("/api/v1/revisions/policy/:policy", auth(api.handleRevisionsGetByPolicy))

	router.POST("/api/v1/state/enforce/noop/:noop", auth(api.handleStateEnforce))

	// return aptomi version
	router.GET("/version", api.handleVersion)
	router.GET("/api/v1/version", api.handleVersion)
}
