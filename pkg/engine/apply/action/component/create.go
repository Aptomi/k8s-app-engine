package component

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
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
func (a *CreateAction) Apply(context *action.Context) (errResult error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}

		action.CollectMetricsFor(a, start, errResult)
	}()

	context.EventLog.NewEntry().Debugf("Creating component instance: %s", a.ComponentKey)

	// deploy to cloud
	instance, err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("unable to deploy component instance '%s': %s", a.ComponentKey, err)
	}

	// update actual state
	return context.ActualStateUpdater.CreateComponentInstance(instance)
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

func (a *CreateAction) processDeployment(context *action.Context) (*resolve.ComponentInstance, error) {
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	if instance == nil {
		panic(fmt.Sprintf("component instance not found in desired state: %s", a.ComponentKey))
	}

	serviceObj, err := context.DesiredPolicy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil {
		// If this is a service instance, do nothing and proceed to object creation
		return instance, nil
	}

	if component.Code == nil {
		// If this is not a code component, do nothing and proceed to object creation
		return instance, nil
	}

	// Instantiate code component
	context.EventLog.NewEntry().Infof("Deploying new component instance: %s", instance.GetKey())

	clusterObj, err := context.DesiredPolicy.GetObject(lang.ClusterObject.Kind, instance.Metadata.Key.ClusterName, instance.Metadata.Key.ClusterNameSpace)
	if err != nil {
		return nil, err
	}
	if clusterObj == nil {
		return nil, fmt.Errorf("cluster '%s/%s' in not present in policy", instance.Metadata.Key.ClusterNameSpace, instance.Metadata.Key.ClusterName)
	}
	cluster := clusterObj.(*lang.Cluster)

	p, err := context.Plugins.ForCodeType(cluster, component.Code.Type)
	if err != nil {
		return nil, err
	}

	return instance, p.Create(
		&plugin.CodePluginInvocationParams{
			DeployName:   instance.GetDeployName(),
			Params:       instance.CalculatedCodeParams,
			PluginParams: map[string]string{plugin.ParamTargetSuffix: instance.Metadata.Key.TargetSuffix},
			EventLog:     context.EventLog,
		},
	)
}
