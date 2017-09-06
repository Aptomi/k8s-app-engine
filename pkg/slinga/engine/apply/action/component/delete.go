package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type DeleteAction struct {
	object.Metadata
	*action.Base

	ComponentKey string
}

func NewDeleteAction(componentKey string) *DeleteAction {
	return &DeleteAction{
		Metadata:     object.Metadata{}, // TODO: initialize
		Base:         action.NewBase(),
		ComponentKey: componentKey,
	}
}

func (componentDelete *DeleteAction) Apply(context *action.Context) error {
	// delete from cloud
	err := componentDelete.processDeployment(context)
	if err != nil {
		context.EventLog.LogError(err)
		return fmt.Errorf("Errors while deleting component '%s': %s", componentDelete.ComponentKey, err)
	}

	// update actual state
	componentDelete.updateActualState(context)
	return nil
}

func (componentDelete *DeleteAction) updateActualState(context *action.Context) {
	// delete component from the actual state
	delete(context.ActualState.ComponentInstanceMap, componentDelete.ComponentKey)
}

func (componentDelete *DeleteAction) processDeployment(context *action.Context) error {
	instance := context.ActualState.ComponentInstanceMap[componentDelete.ComponentKey]
	component := context.ActualPolicy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Instantiate component
	context.EventLog.WithFields(eventlog.Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Destructing a running component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		clusterName, ok := instance.CalculatedCodeParams["cluster"].(string)
		if !ok {
			return fmt.Errorf("No cluster specified in code params, component instance: %v", componentDelete.ComponentKey)
		}

		cluster, ok := context.DesiredPolicy.Clusters[clusterName]
		if !ok {
			return fmt.Errorf("Can't find cluster in policy: %s", clusterName)
		}

		plugin, err := context.Plugins.GetDeployPlugin(component.Code.Type)
		if err != nil {
			return err
		}

		err = plugin.Destroy(cluster, componentDelete.ComponentKey, instance.CalculatedCodeParams, context.EventLog)
		if err != nil {
			return err
		}
	}

	return nil
}
