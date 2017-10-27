package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"time"
)

// CreateActionObject is an informational data structure with Kind and Constructor for the action
var CreateActionObject = &object.Info{
	Kind:        "action-component-create",
	Constructor: func() object.Base { return &CreateAction{} },
}

// CreateAction is a action which gets called when a new component needs to be instantiated (i.e. new instance of code to be deployed to the cloud)
type CreateAction struct {
	*action.Metadata
	ComponentKey string
}

// NewCreateAction creates new CreateAction
func NewCreateAction(revision object.Generation, componentKey string) *CreateAction {
	return &CreateAction{
		Metadata:     action.NewMetadata(revision, CreateActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *CreateAction) Apply(context *action.Context) error {
	// deploy to cloud
	err := a.processDeployment(context)
	if err != nil {
		context.EventLog.LogError(err)
		return fmt.Errorf("error while creating component '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return a.updateActualState(context)
}

func (a *CreateAction) updateActualState(context *action.Context) error {
	// get instance from desired state
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]

	// update creation and update times
	instance.UpdateTimes(time.Now(), time.Now())

	// copy it over to the actual state
	context.ActualState.ComponentInstanceMap[a.ComponentKey] = instance
	err := context.ActualStateUpdater.Create(instance)
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}

func (a *CreateAction) processDeployment(context *action.Context) error {
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
	}).Info("Deploying new component instance: " + instance.GetKey())

	clusterName, ok := instance.CalculatedCodeParams[lang.LabelCluster].(string)
	if !ok {
		return fmt.Errorf("no cluster specified in code params, component instance: %v", a.ComponentKey)
	}

	clusterObj, err := context.DesiredPolicy.GetObject(lang.ClusterObject.Kind, clusterName, object.SystemNS)
	if err != nil {
		return err
	}
	if clusterObj == nil {
		return fmt.Errorf("can't find cluster in policy: %s", clusterName)
	}

	plugin, err := context.Plugins.GetDeployPlugin(component.Code.Type)
	if err != nil {
		return err
	}

	return plugin.Create(clusterObj.(*lang.Cluster), instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
}
