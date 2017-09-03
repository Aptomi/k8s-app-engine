package actions

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"time"
)

type ComponentUpdate struct {
	object.Metadata
	*BaseAction

	ComponentKey string
}

func NewComponentUpdateAction(componentKey string) *ComponentUpdate {
	return &ComponentUpdate{
		Metadata:     object.Metadata{}, // TODO: initialize
		BaseAction:   NewComponentBaseAction(),
		ComponentKey: componentKey,
	}
}

func (componentUpdate *ComponentUpdate) Apply(context *ActionContext) error {
	// update in the cloud
	err := componentUpdate.processDeployment(context)
	if err != nil {
		return fmt.Errorf("Errors while updating component '%s': %s", componentUpdate.ComponentKey, err)
	}

	// update actual state
	componentUpdate.updateActualState(context)
	return nil
}

func (componentUpdate *ComponentUpdate) updateActualState(context *ActionContext) {
	// preserve previous creation date before overwriting
	prevCreatedOn := context.ActualState.ComponentInstanceMap[componentUpdate.ComponentKey].CreatedOn
	instance := context.DesiredState.ComponentInstanceMap[componentUpdate.ComponentKey]
	context.ActualState.ComponentInstanceMap[componentUpdate.ComponentKey] = instance
	instance.UpdateTimes(prevCreatedOn, time.Now())
}

func (componentUpdate *ComponentUpdate) processDeployment(context *ActionContext) error {
	instance := context.DesiredState.ComponentInstanceMap[componentUpdate.ComponentKey]
	component := context.DesiredPolicy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Instantiate component
	context.EventLog.WithFields(eventlog.Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Updating a running component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		clusterName, ok := instance.CalculatedCodeParams["cluster"].(string)
		if !ok {
			return fmt.Errorf("No cluster specified in code params, component instance: %v", instance.Key)
		}

		cluster, ok := context.DesiredPolicy.Clusters[clusterName]
		if !ok {
			return fmt.Errorf("No specified cluster in policy: %s", clusterName)
		}

		plugin, err := context.Plugins.GetDeployPlugin(component.Code.Type)
		if err != nil {
			return err
		}

		err = plugin.Update(cluster, componentUpdate.ComponentKey, instance.CalculatedCodeParams, context.EventLog)
		if err != nil {
			return err
		}
	}

	return nil
}
