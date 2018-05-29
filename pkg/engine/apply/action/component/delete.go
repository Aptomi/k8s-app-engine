package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
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
	Params       util.NestedParameterMap
}

// NewDeleteAction creates new DeleteAction
func NewDeleteAction(componentKey string, params util.NestedParameterMap) *DeleteAction {
	return &DeleteAction{
		TypeKind:     DeleteActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(DeleteActionObject.Kind, componentKey),
		ComponentKey: componentKey,
		Params:       params,
	}
}

// Apply applies the action
func (a *DeleteAction) Apply(context *action.Context) error {
	// delete from cloud
	err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("unable to delete component instance '%s': %s", a.ComponentKey, err)
	}

	// delete from the actual state
	return deleteComponentFromActualState(a.ComponentKey, context)
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *DeleteAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":   a.Kind,
		"key":    a.ComponentKey,
		"params": a.Params,
		"pretty": fmt.Sprintf("[-] %s", a.ComponentKey),
	}
}

func (a *DeleteAction) processDeployment(context *action.Context) error {
	instance := context.ActualState.ComponentInstanceMap[a.ComponentKey]
	serviceObj, err := context.DesiredPolicy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing and report successful deletion
		return nil
	}

	if component.Code == nil {
		// This is a non-code component. Do nothing and report successful deletion
		return nil
	}

	context.EventLog.NewEntry().Infof("Destructing a running component instance: %s", instance.GetKey())

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

	return plugin.Destroy(instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
}
