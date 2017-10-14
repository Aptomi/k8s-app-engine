package api

import (
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (a *api) handleAdminStoreDump(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.store.Object().Dump(w)
	if err != nil {
		logrus.Panicf("Error while dumping db to response: %s", err)
	}
}
