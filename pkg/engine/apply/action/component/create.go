package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// CreateActionObject is an informational data structure with Kind and Constructor for the action
var CreateActionObject = &runtime.Info{
	Kind:        "action-component-create",
	Constructor: func() runtime.Object { return &CreateAction{} },
}

// CreateAction is a action which gets called when a new component needs to be instantiated (i.e. new instance of code to be deployed to the cloud)
type CreateAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
}

// NewCreateAction creates new CreateAction
func NewCreateAction(componentKey string) *CreateAction {
	return &CreateAction{
		TypeKind:     CreateActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(CreateActionObject.Kind, componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *CreateAction) Apply(context *action.Context) error {
	// deploy to cloud
	err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("unable to deploy component instance '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return updateActualStateFromDesired(a.ComponentKey, context, true, true, true)
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

	return plugin.Create(instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
}
