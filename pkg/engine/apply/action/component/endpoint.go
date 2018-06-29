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

// EndpointsAction is a action which gets called when a new component changed (created or updated) and endpoints should be updated
type EndpointsAction struct {
	*action.Metadata
	ComponentKey string
}

// NewEndpointsAction creates new EndpointsAction
func NewEndpointsAction(componentKey string) *EndpointsAction {
	return &EndpointsAction{
		Metadata:     action.NewMetadata("action-component-endpoints", componentKey),
		ComponentKey: componentKey,
	}
}

// Apply applies the action
func (a *EndpointsAction) Apply(context *action.Context) (errResult error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}

		action.CollectMetricsFor(a, start, errResult)
	}()

	context.EventLog.NewEntry().Infof("Getting endpoints for component instance: %s", a.ComponentKey)

	// fetch component endpoints and store them in component instance (actual state)
	instance, endpoints, err := a.processEndpoints(context)
	if err != nil {
		return fmt.Errorf("unable to get endpoints for component instance '%s': %s", a.ComponentKey, err)
	}

	// update component endpoints in actual state
	return context.ActualStateUpdater.UpdateComponentInstance(instance.GetKey(), func(obj *resolve.ComponentInstance) {
		obj.EndpointsUpToDate = true
		obj.Endpoints = endpoints
	})
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *EndpointsAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":       a.Kind,
		"key":        a.ComponentKey,
		"pretty":     fmt.Sprintf("[@] %s", a.ComponentKey),
		"prettyOmit": "true", // do not print endpoint lines in pretty output
	}
}

func (a *EndpointsAction) processEndpoints(context *action.Context) (*resolve.ComponentInstance, map[string]string, error) {
	instance := context.ActualStateUpdater.GetComponentInstance(a.ComponentKey)
	if instance == nil {
		return nil, nil, fmt.Errorf("component instance not found in actual state: %s", a.ComponentKey)
	}

	bundleObj, err := context.DesiredPolicy.GetObject(lang.TypeBundle.Kind, instance.Metadata.Key.BundleName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, nil, err
	}
	component := bundleObj.(*lang.Bundle).GetComponentsMap()[instance.Metadata.Key.ComponentName] // nolint: errcheck

	if component == nil {
		return nil, nil, fmt.Errorf("retrieving endpoints for bundle instance is not supported")
	}

	// endpoints could be calculated only for components with code
	if component.Code == nil {
		return nil, nil, fmt.Errorf("retrieving endpoints for non-code components is not supported")
	}

	clusterObj, err := context.DesiredPolicy.GetObject(lang.TypeCluster.Kind, instance.Metadata.Key.ClusterName, instance.Metadata.Key.ClusterNameSpace)
	if err != nil {
		return nil, nil, err
	}
	if clusterObj == nil {
		return nil, nil, fmt.Errorf("cluster '%s/%s' in not present in policy", instance.Metadata.Key.ClusterNameSpace, instance.Metadata.Key.ClusterName)
	}
	cluster := clusterObj.(*lang.Cluster) // nolint: errcheck

	p, err := context.Plugins.ForCodeType(cluster, component.Code.Type)
	if err != nil {
		return nil, nil, err
	}

	endpoints, err := p.Endpoints(
		&plugin.CodePluginInvocationParams{
			DeployName:   instance.GetDeployName(),
			Params:       instance.CalculatedCodeParams,
			PluginParams: map[string]string{plugin.ParamTargetSuffix: instance.Metadata.Key.TargetSuffix},
			EventLog:     context.EventLog,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	context.EventLog.NewEntry().Infof("Received %d endpoints for component instance: %s", len(endpoints), a.ComponentKey)

	return instance, endpoints, err
}
