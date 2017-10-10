package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"time"
)

// UpdateActionObject is an informational data structure with Kind and Constructor for the action
var UpdateActionObject = &object.Info{
	Kind:        "action-component-update",
	Constructor: func() object.Base { return &DeleteAction{} },
}

// UpdateAction is a action which gets called when an existing component needs to be updated (i.e. parameters of a running code instance need to be changed in the cloud)
type UpdateAction struct {
	*action.Metadata
	ComponentKey string
}

// NewUpdateAction creates new UpdateAction
func NewUpdateAction(revision object.Generation, componentKey string) *UpdateAction {
	return &UpdateAction{
		Metadata:     action.NewMetadata(revision, UpdateActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *UpdateAction) Apply(context *action.Context) error {
	// update in the cloud
	err := a.processDeployment(context)
	if err != nil {
		context.EventLog.LogError(err)
		return fmt.Errorf("Errors while updating component '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return a.updateActualState(context)
}

func (a *UpdateAction) updateActualState(context *action.Context) error {
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

func (a *UpdateAction) processDeployment(context *action.Context) error {
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

	if component.Code == nil {
		return nil
	}

	context.EventLog.WithFields(event.Fields{
		"componentKey": instance.Metadata.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Updating a running component instance: " + instance.GetKey())

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

	err = plugin.Update(clusterObj.(*lang.Cluster), instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
	if err != nil {
		return err
	}

	return nil
}
