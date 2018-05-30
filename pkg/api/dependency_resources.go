package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type dependencyResourcesWrapper struct {
	Resources plugin.Resources
}

func (g *dependencyResourcesWrapper) GetKind() string {
	return "dependencyResources"
}

func (api *coreAPI) handleDependencyResourcesGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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

	plugins := api.pluginRegistryFactory()
	depKey := runtime.KeyForStorable(dependency)
	resources := make(plugin.Resources)
	for _, instance := range actualState.ComponentInstanceMap {
		if _, ok := instance.DependencyKeys[depKey]; ok {
			codePlugin, pluginErr := pluginForComponentInstance(instance, policy, plugins)
			if pluginErr != nil {
				panic(fmt.Sprintf("Can't get plugin for component instance %s: %s", instance.GetKey(), pluginErr))
			}
			if codePlugin == nil {
				continue
			}

			instanceResources, resErr := codePlugin.Resources(instance.GetDeployName(), instance.CalculatedCodeParams, event.NewLog(logrus.WarnLevel, "resources"))
			if resErr != nil {
				panic(fmt.Sprintf("Error while getting deployment resources for component instance %s: %s", instance.GetKey(), resErr))
			}

			resources.Merge(instanceResources)
		}
	}

	api.contentType.WriteOne(writer, request, &dependencyResourcesWrapper{resources})
}

func pluginForComponentInstance(instance *resolve.ComponentInstance, policy *lang.Policy, plugins plugin.Registry) (plugin.CodePlugin, error) {
	serviceObj, err := policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil || component.Code == nil {
		return nil, nil
	}

	clusterName := instance.GetCluster()
	if len(clusterName) <= 0 {
		return nil, fmt.Errorf("component instance does not have cluster assigned: %s", instance.GetKey())
	}

	clusterObj, err := policy.GetObject(lang.ClusterObject.Kind, clusterName, runtime.SystemNS)
	if err != nil {
		return nil, err
	}
	if clusterObj == nil {
		return nil, fmt.Errorf("can't find cluster in policy: %s", clusterName)
	}
	cluster := clusterObj.(*lang.Cluster)

	return plugins.ForCodeType(cluster, component.Code.Type)
}
