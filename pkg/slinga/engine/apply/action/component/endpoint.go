package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"time"
)

// EndpointsActionObject is an informational data structure with Kind and Constructor for the action
var EndpointsActionObject = &object.Info{
	Kind:        "action-component-create",
	Constructor: func() object.Base { return &CreateAction{} },
}

// EndpointsAction is a action which gets called when a new component changed (created or updated) and endpoints should be updated
type EndpointsAction struct {
	// Key is the revision id and action id pair
	*action.Metadata
	ComponentKey string
}

// NewEndpointsAction creates new EndpointsAction
func NewEndpointsAction(revision object.Generation, componentKey string) *EndpointsAction {
	return &EndpointsAction{
		Metadata:     action.NewMetadata(revision, EndpointsActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *EndpointsAction) Apply(context *action.Context) error {
	err := a.processEndpoints(context)
	if err != nil {
		context.EventLog.LogError(err)
		return fmt.Errorf("Errors while getting endpoints '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return a.updateActualState(context)
}

func (a *EndpointsAction) updateActualState(context *action.Context) error {
	// preserve previous creation date before overwriting
	prevCreatedOn := context.ActualState.ComponentInstanceMap[a.ComponentKey].CreatedOn
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	instance.UpdateTimes(prevCreatedOn, time.Now())

	context.ActualState.ComponentInstanceMap[a.ComponentKey] = instance
	err := context.ActualStateUpdater.Update(instance)
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}

func (a *EndpointsAction) processEndpoints(context *action.Context) error {
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	serviceObj, err := context.DesiredPolicy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
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

	clusterName, ok := instance.CalculatedCodeParams[lang.LabelCluster].(string)
	if !ok {
		return fmt.Errorf("No cluster specified in code params, component instance: %v", a.ComponentKey)
	}

	clusterObj, err := context.DesiredPolicy.GetObject(lang.ClusterObject.Kind, clusterName, object.SystemNS)
	if err != nil {
		return err
	}
	if clusterObj == nil {
		return fmt.Errorf("Can't find cluster in policy: %s", clusterName)
	}

	plugin, err := context.Plugins.GetDeployPlugin(component.Code.Type)
	if err != nil {
		return err
	}

	endpoints, err := plugin.Endpoints(clusterObj.(*lang.Cluster), instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
	if err != nil {
		return err
	}

	instance.Endpoints = endpoints

	return nil
}
