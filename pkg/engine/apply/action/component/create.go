package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
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
	Params       util.NestedParameterMap
}

// NewCreateAction creates new CreateAction
func NewCreateAction(componentKey string, params util.NestedParameterMap) *CreateAction {
	return &CreateAction{
		TypeKind:     CreateActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(CreateActionObject.Kind, componentKey),
		ComponentKey: componentKey,
		Params:       params,
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
	return createComponentInActualState(a.ComponentKey, context)
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *CreateAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":   a.Kind,
		"key":    a.ComponentKey,
		"params": a.Params,
		"pretty": fmt.Sprintf("[+] %s", a.ComponentKey),
	}
}

func (a *CreateAction) processDeployment(context *action.Context) error {
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	serviceObj, err := context.DesiredPolicy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil {
		// If this is a service instance, do nothing
		return nil
	}

	if component.Code == nil {
		// If this is not a code component, do nothing
		return nil
	}

	// Instantiate code component
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
