package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"time"
)

type UpdateAction struct {
	object.Metadata
	*action.Base

	ComponentKey string
}

func NewUpdateAction(componentKey string) *UpdateAction {
	return &UpdateAction{
		Metadata:     object.Metadata{}, // TODO: initialize
		Base:         action.NewBase(),
		ComponentKey: componentKey,
	}
}

func (a *UpdateAction) Apply(context *action.Context) error {
	// update in the cloud
	err := a.processDeployment(context)
	if err != nil {
		context.EventLog.LogError(err)
		return fmt.Errorf("Errors while updating component '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	a.updateActualState(context)
	return nil
}

func (a *UpdateAction) updateActualState(context *action.Context) {
	// preserve previous creation date before overwriting
	prevCreatedOn := context.ActualState.ComponentInstanceMap[a.ComponentKey].CreatedOn
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	context.ActualState.ComponentInstanceMap[a.ComponentKey] = instance
	instance.UpdateTimes(prevCreatedOn, time.Now())
}

func (a *UpdateAction) processDeployment(context *action.Context) error {
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	serviceComponent := context.DesiredPolicy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if serviceComponent == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Instantiate component
	context.EventLog.WithFields(eventlog.Fields{
		"componentKey": instance.Key,
		"component":    serviceComponent.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Updating a running component instance: " + instance.Key.GetKey())

	if serviceComponent.Code != nil {
		clusterName, ok := instance.CalculatedCodeParams["cluster"].(string)
		if !ok {
			return fmt.Errorf("No cluster specified in code params, component instance: %v", a.ComponentKey)
		}

		cluster, ok := context.DesiredPolicy.Clusters[clusterName]
		if !ok {
			return fmt.Errorf("Can't find cluster in policy: %s", clusterName)
		}

		plugin, err := context.Plugins.GetDeployPlugin(serviceComponent.Code.Type)
		if err != nil {
			return err
		}

		err = plugin.Update(cluster, a.ComponentKey, instance.CalculatedCodeParams, context.EventLog)
		if err != nil {
			return err
		}
	}

	return nil
}
