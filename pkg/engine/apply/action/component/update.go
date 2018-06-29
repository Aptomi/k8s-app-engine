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

// UpdateAction is a action which gets called when an existing component needs to be updated (i.e. parameters of a running code instance need to be changed in the cloud)
type UpdateAction struct {
	*action.Metadata
	ComponentKey string
	ParamsBefore util.NestedParameterMap
	Params       util.NestedParameterMap
}

// NewUpdateAction creates new UpdateAction
func NewUpdateAction(componentKey string, paramsBefore util.NestedParameterMap, params util.NestedParameterMap) *UpdateAction {
	return &UpdateAction{
		Metadata:     action.NewMetadata("action-component-update", componentKey),
		ComponentKey: componentKey,
		ParamsBefore: paramsBefore,
		Params:       params,
	}
}

// Apply applies the action
func (a *UpdateAction) Apply(context *action.Context) (errResult error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}

		action.CollectMetricsFor(a, start, errResult)
	}()

	context.EventLog.NewEntry().Debugf("Updating component instance: %s", a.ComponentKey)

	// update in the cloud
	instance, err := a.processDeployment(context)
	if err != nil {
		return fmt.Errorf("unable to update component instance '%s': %s", a.ComponentKey, err)
	}

	// update component instance code params in actual state
	if instance.CalculatedCodeParams != nil {
		return context.ActualStateUpdater.UpdateComponentInstance(instance.GetKey(), func(obj *resolve.ComponentInstance) {
			obj.EndpointsUpToDate = false // invalidate endpoints, so we retrieve them again later
			obj.CalculatedCodeParams = instance.CalculatedCodeParams
		})
	}

	return nil
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

func (a *UpdateAction) processDeployment(context *action.Context) (*resolve.ComponentInstance, error) {
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	if instance == nil {
		return nil, fmt.Errorf("component instance not found desired state: %s", a.ComponentKey)
	}

	bundleObj, err := context.DesiredPolicy.GetObject(lang.BundleType.Kind, instance.Metadata.Key.BundleName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, err
	}
	component := bundleObj.(*lang.Bundle).GetComponentsMap()[instance.Metadata.Key.ComponentName] // nolint: errcheck

	if component == nil {
		// This is a bundle instance. Do nothing and proceed with object update
		return instance, nil
	}

	if component.Code == nil {
		// This is a bundle instance. Do nothing and proceed with object update
		return instance, nil
	}

	context.EventLog.NewEntry().Infof("Updating a running component instance: %s ", instance.GetKey())

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

	return instance, p.Update(
		&plugin.CodePluginInvocationParams{
			DeployName:   instance.GetDeployName(),
			Params:       instance.CalculatedCodeParams,
			PluginParams: map[string]string{plugin.ParamTargetSuffix: instance.Metadata.Key.TargetSuffix},
			EventLog:     context.EventLog,
		},
	)
}
