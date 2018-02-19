package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// UpdateActionObject is an informational data structure with Kind and Constructor for the action
var UpdateActionObject = &runtime.Info{
	Kind:        "action-component-update",
	Constructor: func() runtime.Object { return &DeleteAction{} },
}

// UpdateAction is a action which gets called when an existing component needs to be updated (i.e. parameters of a running code instance need to be changed in the cloud)
type UpdateAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
}

// NewUpdateAction creates new UpdateAction
func NewUpdateAction(componentKey string) *UpdateAction {
	return &UpdateAction{
		TypeKind:     UpdateActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(UpdateActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *UpdateAction) Apply(context *action.Context) error {
	// update in the cloud
	err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("error while updating component '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return updateActualStateFromDesired(a.ComponentKey, context, false, true, false)
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
		return fmt.Errorf("no cluster specified in code params, component instance: %v", a.ComponentKey)
	}

	clusterObj, err := context.DesiredPolicy.GetObject(lang.ClusterObject.Kind, clusterName, runtime.SystemNS)
	if err != nil {
		return err
	}
	if clusterObj == nil {
		return fmt.Errorf("can't find cluster in policy: %s", clusterName)
	}
	cluster := clusterObj.(*lang.Cluster)

	plugin, err := context.Plugins.ForCodeType(cluster, component.Code.Type)
	if err != nil {
		return err
	}

	return plugin.Update(instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
}
