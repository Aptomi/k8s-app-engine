package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// EndpointsActionObject is an informational data structure with Kind and Constructor for the action
var EndpointsActionObject = &runtime.Info{
	Kind:        "action-component-endpoints",
	Constructor: func() runtime.Object { return &EndpointsAction{} },
}

// EndpointsAction is a action which gets called when a new component changed (created or updated) and endpoints should be updated
type EndpointsAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
}

// NewEndpointsAction creates new EndpointsAction
func NewEndpointsAction(componentKey string) *EndpointsAction {
	return &EndpointsAction{
		TypeKind:     EndpointsActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(EndpointsActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *EndpointsAction) Apply(context *action.Context) error {
	// if component for some reason doesn't exist in actual state, report an error
	if context.ActualState.ComponentInstanceMap[a.ComponentKey] == nil {
		return fmt.Errorf("unable to get endpoints for component instance '%s': it doesn't exist in actual state", a.ComponentKey)
	}

	err := a.processEndpoints(context)
	if err != nil {
		return fmt.Errorf("unable to get endpoints for component instance '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return updateComponentInActualState(a.ComponentKey, context)
}

func (a *EndpointsAction) processEndpoints(context *action.Context) error {
	instance := context.ActualState.ComponentInstanceMap[a.ComponentKey]
	serviceObj, err := context.DesiredPolicy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil {
		return fmt.Errorf("retrieving endpoints for service instance is not supported")
	}

	// endpoints could be calculated only for components with code
	if component.Code == nil {
		return nil
	}

	context.EventLog.WithFields(event.Fields{
		"componentKey": instance.Metadata.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Getting endpoints for component instance: " + instance.GetKey())

	clusterName := instance.GetCluster()
	if len(clusterName) <= 0 {
		return fmt.Errorf("policy doesn't specify deployment target for component instance")
	}

	clusterObj, err := context.DesiredPolicy.GetObject(lang.ClusterObject.Kind, clusterName, runtime.SystemNS)
	if err != nil {
		return err
	}
	if clusterObj == nil {
		return fmt.Errorf("cluster '%s' in not present in policy", clusterName)
	}
	cluster := clusterObj.(*lang.Cluster)

	plugin, err := context.Plugins.ForCodeType(cluster, component.Code.Type)
	if err != nil {
		return err
	}

	endpoints, err := plugin.Endpoints(instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
	if err != nil {
		return err
	}

	instance.Endpoints = endpoints

	return nil
}
