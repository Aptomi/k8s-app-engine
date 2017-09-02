package deployment

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/base"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type DeployerPlugin struct {
	*base.BasePlugin
}

func (deployer *DeployerPlugin) GetCustomApplyProgressLength() int {
	return 0
}

func (deployer *DeployerPlugin) OnApplyComponentInstanceCreate(key string) error {
	instance := deployer.Desired.Resolution.ComponentInstanceMap[key]
	component := deployer.Desired.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Instantiate component
	deployer.EventLog.WithFields(Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Deploying new component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		codeExecutor, err := GetCodeExecutor(
			component.Code,
			instance.Key.GetKey(),
			instance.CalculatedCodeParams,
			deployer.Desired.Policy.Clusters,
			deployer.EventLog,
		)
		if err != nil {
			return err
		}

		err = codeExecutor.Install()
		if err != nil {
			return err
		}
	}

	return nil
}

func (deployer *DeployerPlugin) OnApplyComponentInstanceUpdate(key string) error {
	instance := deployer.Desired.Resolution.ComponentInstanceMap[key]
	component := deployer.Desired.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Update component
	deployer.EventLog.WithFields(Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Updating a running component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		codeExecutor, err := GetCodeExecutor(
			component.Code,
			instance.Key.GetKey(),
			instance.CalculatedCodeParams,
			deployer.Desired.Policy.Clusters,
			deployer.EventLog,
		)
		if err != nil {
			return err
		}
		err = codeExecutor.Update()
		if err != nil {
			return err
		}

	}
	return nil
}

func (deployer *DeployerPlugin) OnApplyComponentInstanceDelete(key string) error {
	instance := deployer.Actual.Resolution.ComponentInstanceMap[key]
	component := deployer.Actual.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]
	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Delete component
	deployer.EventLog.WithFields(Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("Destructing a running component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		codeExecutor, err := GetCodeExecutor(
			component.Code,
			instance.Key.GetKey(),
			instance.CalculatedCodeParams,
			deployer.Actual.Policy.Clusters,
			deployer.EventLog,
		)
		if err != nil {
			return err
		}
		err = codeExecutor.Destroy()
		if err != nil {
			return err
		}
	}
	return nil
}

func (deployer *DeployerPlugin) OnApplyCustom(progress progress.ProgressIndicator) error {
	return nil
}
