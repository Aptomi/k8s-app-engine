package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// EndpointsObject is an informational data structure with Kind and Constructor for Endpoints
var EndpointsObject = &runtime.Info{
	Kind:        "endpoints",
	Constructor: func() runtime.Object { return &Endpoints{} },
}

// Endpoints object represents
type Endpoints struct {
	runtime.TypeKind `yaml:",inline"`
	List             map[string]map[string]string
}

func (api *coreAPI) handleEndpointsGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	endpoints := make(map[string]map[string]string)
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("Can't load actual state to get endpoints: %s", err))
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
