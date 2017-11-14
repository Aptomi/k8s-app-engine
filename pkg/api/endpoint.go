package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
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
	dependencyNamespace := params.ByName("ns")
	dependencyName := params.ByName("name")
	filterEnabled := len(dependencyNamespace) > 0 && len(dependencyName) > 0

	endpoints := make(map[string]map[string]string)
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("Can't load actual state to get endpoints: %s", err))
	}
	for _, instance := range actualState.ComponentInstanceMap {
		if len(instance.Endpoints) > 0 {
			// filter by dependency, if/as needed
			add := true
			if filterEnabled {
				add = false
				// ideally we want to retrieve the corresponding dependency from policy, but for now let's just do
				// string-based checks (because key for storable objects is namespace/kind/name)
				for key := range instance.DependencyKeys {
					if strings.HasPrefix(key, dependencyNamespace) && strings.HasSuffix(key, dependencyName) {
						add = true
						break
					}
				}
			}

			// collect endpoints
			if add {
				endpoints[instance.GetName()] = instance.Endpoints
			}
		}
	}

	api.contentType.WriteOne(writer, request, &Endpoints{
		TypeKind: EndpointsObject.GetTypeKind(),
		List:     endpoints,
	})
}
