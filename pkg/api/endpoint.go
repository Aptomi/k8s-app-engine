package api

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var EndpointsObject = &runtime.Info{
	Kind:        "endpoints",
	Constructor: func() runtime.Object { return &Endpoints{} },
}

type Endpoints struct {
	runtime.TypeKind `yaml:",inline"`
	List             map[string]map[string]string
}

func (api *coreApi) handleEndpointsGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	endpoints := make(map[string]map[string]string)
	actualState, err := api.store.GetActualState()
	if err != nil {
		log.Panicf("Can't load actual state to get endpoints: %s", err)
	}
	for _, instance := range actualState.ComponentInstanceMap {
		if len(instance.Endpoints) > 0 {
			endpoints[instance.GetName()] = instance.Endpoints
		}
	}

	api.contentType.Write(writer, request, &Endpoints{
		TypeKind: EndpointsObject.GetTypeKind(),
		List:     endpoints,
	})
}
