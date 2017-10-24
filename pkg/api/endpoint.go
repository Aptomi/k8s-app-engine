package api

import (
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (a *api) handleEndpointsShow(r *http.Request, p httprouter.Params) Response {
	endpoints := make(map[string]map[string]string)
	actualState, err := a.store.GetActualState()
	if err != nil {
		log.Panicf("Can't load actual state to get endpoints: %s", err)
	}
	for _, instance := range actualState.ComponentInstanceMap {
		if len(instance.Endpoints) > 0 {
			endpoints[instance.GetName()] = instance.Endpoints
		}
	}

	return endpoints
}
