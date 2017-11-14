package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type dependencyStatusWrapper struct {
	Data interface{}
}

func (g *dependencyStatusWrapper) GetKind() string {
	return "dependencyStatus"
}

func (api *coreAPI) handleDependencyStatusGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := runtime.LastGen
	policy, _, err := api.store.GetPolicy(gen)
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	ns := params.ByName("ns")
	kind := lang.DependencyObject.Kind
	name := params.ByName("name")

	obj, err := policy.GetObject(kind, name, ns)
	if err != nil {
		panic(fmt.Sprintf("error while getting object %s/%s/%s in policy #%s", ns, kind, name, gen))
	}
	if obj == nil {
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusNotFound)
	}

	// once dependency is loaded, we need to find its state in the actual state
	dependency := obj.(*lang.Dependency)
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("Can't load actual state to get endpoints: %s", err))
	}

	var status string
	key := runtime.KeyForStorable(dependency)

	foundRefs := false
	for _, instance := range actualState.ComponentInstanceMap {
		if _, ok := instance.DependencyKeys[key]; ok {
			foundRefs = true
			break
		}
	}
	if foundRefs {
		status = "Deployed"
	} else {
		status = "Not Deployed"
	}

	api.contentType.WriteOne(writer, request, &dependencyStatusWrapper{Data: status})
}
