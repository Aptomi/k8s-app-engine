package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// DeleteActionObject is an informational data structure with Kind and Constructor for the action
var DeleteActionObject = &runtime.Info{
	Kind:        "action-component-delete",
	Constructor: func() runtime.Object { return &DeleteAction{} },
}

// DeleteAction is a action which gets called when an existing component needs to be destroyed (i.e. existing instance of code needs to be terminated in the cloud)
type DeleteAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
}

// NewDeleteAction creates new DeleteAction
func NewDeleteAction(componentKey string) *DeleteAction {
	return &DeleteAction{
		TypeKind:     DeleteActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(DeleteActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *DeleteAction) Apply(context *action.Context) error {
	// delete from cloud
	err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("error while deleting component '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return deleteComponentFromActualState(a.ComponentKey, context)
}

func (a *DeleteAction) processDeployment(context *action.Context) error {
	instance := context.ActualState.ComponentInstanceMap[a.ComponentKey]
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
	}).Info("Destructing a running component instance: " + instance.GetKey())

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

	return plugin.Destroy(instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
}
