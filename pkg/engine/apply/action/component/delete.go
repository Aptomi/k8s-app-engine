package component

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/util"
)

// DeleteAction is a action which gets called when an existing component needs to be destroyed (i.e. existing instance of code needs to be terminated in the cloud)
type DeleteAction struct {
	*action.Metadata
	ComponentKey string
	Params       util.NestedParameterMap
}

// NewDeleteAction creates new DeleteAction
func NewDeleteAction(componentKey string, params util.NestedParameterMap) *DeleteAction {
	return &DeleteAction{
		Metadata:     action.NewMetadata("action-component-delete", componentKey),
		ComponentKey: componentKey,
		Params:       params,
	}
}

// Apply applies the action
func (a *DeleteAction) Apply(context *action.Context) (errResult error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}

		action.CollectMetricsFor(a, start, errResult)
	}()

	context.EventLog.NewEntry().Debugf("Deleting component instance: %s", a.ComponentKey)

	// delete from cloud
	instance, err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("unable to delete component instance '%s': %s", a.ComponentKey, err)
	}

	// delete from the actual state
	return context.ActualStateUpdater.DeleteComponentInstance(instance.GetKey())
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

func (a *DeleteAction) processDeployment(context *action.Context) (*resolve.ComponentInstance, error) {
	instance := context.ActualStateUpdater.GetComponentInstance(a.ComponentKey)
	if instance == nil {
		panic(fmt.Sprintf("component instance not found in actual state: %s", a.ComponentKey))
	}

	bundleObj, err := context.DesiredPolicy.GetObject(lang.BundleType.Kind, instance.Metadata.Key.BundleName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, err
	}
	component := bundleObj.(*lang.Bundle).GetComponentsMap()[instance.Metadata.Key.ComponentName] // nolint: errcheck

	if component == nil {
		// This is a bundle instance. Do nothing and proceed with deletion
		return instance, nil
	}

	if component.Code == nil {
		// This is a non-code component. Do nothing and proceed with deletion
		return instance, nil
	}

	context.EventLog.NewEntry().Infof("Destructing a running component instance: %s", instance.GetKey())

	clusterObj, err := context.DesiredPolicy.GetObject(lang.ClusterObject.Kind, instance.Metadata.Key.ClusterName, instance.Metadata.Key.ClusterNameSpace)
	if err != nil {
		return nil, err
	}
	if clusterObj == nil {
		return nil, fmt.Errorf("cluster '%s/%s' in not present in policy", instance.Metadata.Key.ClusterNameSpace, instance.Metadata.Key.ClusterName)
	}
	cluster := clusterObj.(*lang.Cluster) // nolint: errcheck

	p, err := context.Plugins.ForCodeType(cluster, component.Code.Type)
	if err != nil {
		return nil, err
	}

	return instance, p.Destroy(
		&plugin.CodePluginInvocationParams{
			DeployName:   instance.GetDeployName(),
			Params:       instance.CalculatedCodeParams,
			PluginParams: map[string]string{plugin.ParamTargetSuffix: instance.Metadata.Key.TargetSuffix},
			EventLog:     context.EventLog,
		},
	)

}
