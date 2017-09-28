package api

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func ServeAdminStore(router *httprouter.Router, store store.ObjectStore) {
	router.GET("/api/v1/admin/store", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := store.Dump(w)
		if err != nil {
			panic("Error while dumping db to response")
		}
	})
}
