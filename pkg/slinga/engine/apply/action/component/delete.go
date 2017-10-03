package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var DeleteActionObject = &object.Info{
	Kind:        "action-component-delete",
	Constructor: func() object.Base { return &DeleteAction{} },
}

type DeleteAction struct {
	*action.Metadata
	ComponentKey string
}

func NewDeleteAction(revision object.Generation, componentKey string) *DeleteAction {
	return &DeleteAction{
		Metadata:     action.NewMetadata(revision, DeleteActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

func (a *DeleteAction) GetName() string {
	return "Delete component " + a.ComponentKey
}

func (a *DeleteAction) Apply(context *action.Context) error {
	// delete from cloud
	err := a.processDeployment(context)
	if err != nil {
		context.EventLog.LogError(err)
		return fmt.Errorf("Errors while deleting component '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return a.updateActualState(context)
}

func (a *DeleteAction) updateActualState(context *action.Context) error {
	// delete component from the actual state
	delete(context.ActualState.ComponentInstanceMap, a.ComponentKey)
	err := context.ActualStateUpdater.Delete(a.ComponentKey)
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}

func (a *DeleteAction) processDeployment(context *action.Context) error {
	instance := context.ActualState.ComponentInstanceMap[a.ComponentKey]
	serviceObj, err := context.ActualPolicy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Instantiate component
	context.EventLog.WithFields(event.Fields{
		"componentKey": instance.Metadata.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Destructing a running component instance: " + instance.GetKey())

	if component.Code != nil {
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

		err = plugin.Destroy(clusterObj.(*lang.Cluster), a.ComponentKey, instance.CalculatedCodeParams, context.EventLog)
		if err != nil {
			return err
		}
	}

	return nil
}
