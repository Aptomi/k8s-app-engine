package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
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
	ParamsBefore util.NestedParameterMap
	Params       util.NestedParameterMap
}

// NewUpdateAction creates new UpdateAction
func NewUpdateAction(componentKey string, paramsBefore util.NestedParameterMap, params util.NestedParameterMap) *UpdateAction {
	return &UpdateAction{
		TypeKind:     UpdateActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(UpdateActionObject.Kind, componentKey),
		ComponentKey: componentKey,
		ParamsBefore: paramsBefore,
		Params:       params,
	}
}

// Apply applies the action
func (a *UpdateAction) Apply(context *action.Context) error {
	// update in the cloud
	err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("unable to update component instance '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return updateComponentInActualState(a.ComponentKey, context)
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *UpdateAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":         a.Kind,
		"key":          a.ComponentKey,
		"paramsBefore": a.ParamsBefore,
		"params":       a.Params,
		"paramsDiff":   a.ParamsBefore.Diff(a.Params),
		"pretty":       fmt.Sprintf("[*] %s", a.ComponentKey),
	}
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

	context.EventLog.NewEntry().Infof("Updating a running component instance: %s ", instance.GetKey())

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

	err = plugin.Update(instance.GetDeployName(), instance.CalculatedCodeParams, context.EventLog)
	if err != nil {
		return err
	}

	context.ActualState.ComponentInstanceMap[a.ComponentKey].CalculatedCodeParams = instance.CalculatedCodeParams

	return nil
}
