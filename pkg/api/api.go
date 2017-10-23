package api

import (
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/codec"
	"github.com/Aptomi/aptomi/pkg/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/julienschmidt/httprouter"
)

type api struct {
	router       *httprouter.Router
	codec        codec.MarshallerUnmarshaller
	store        store.ServerStore
	externalData *external.Data
}

// Serve initializes everything needed by HTTP API and puts it into the provided http router
func Serve(router *httprouter.Router, store store.ServerStore, externalData *external.Data) {
	catalog := object.NewCatalog().Append(lang.Objects...)
	a := &api{router, yaml.NewCodec(catalog), store, externalData}
	a.serve()
}

func (a *api) serve() {
	a.get("/api/v1/policy", a.handlePolicyShow)
	a.get("/api/v1/policy/gen/:gen", a.handlePolicyShow)
	//a.get("/api/v1/policy/gen/:gen/ns/:namespace", a.handlePolicyShow)
	a.post("/api/v1/policy", a.handlePolicyUpdate)

	a.get("/api/v1/endpoints", a.handleEndpointsShow)

	a.getStream("/api/v1/admin/store", a.handleAdminStoreDump)

	a.get("/version", handleVersion)
}
