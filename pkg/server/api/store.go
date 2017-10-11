package api

import (
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// ServeAdminStore registers admin-level data store viewing handlers in API
func ServeAdminStore(router *httprouter.Router, store store.ServerStore) {
	router.GET("/api/v1/admin/store", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := store.Object().Dump(w)
		if err != nil {
			panic("Error while dumping db to response")
		}
	})
}
