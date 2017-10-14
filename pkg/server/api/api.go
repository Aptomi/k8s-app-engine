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

func New(router *httprouter.Router, store store.ServerStore, externalData *external.Data) *api {
	catalog := object.NewCatalog().Append(lang.Objects...)
	a := &api{router, yaml.NewCodec(catalog), store, externalData}
	a.serve()
	return a
}

func (a *api) serve() {
	a.get("/api/v1/policy", a.handlePolicyShow)
	a.get("/api/v1/policy/gen/:gen", a.handlePolicyShow)
	//a.get("/api/v1/policy/gen/:gen/ns/:namespace", a.handlePolicyShow)
	a.post("/api/v1/policy", a.handlePolicyUpdate)

	a.getStream("/api/v1/admin/store", a.handleAdminStoreDump)

	a.get("/version", handleVersion)
}
