package deployment

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/base"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
)

type DeployerPlugin struct {
	*base.BasePlugin
}

func (deployer *DeployerPlugin) GetApplyProgressLength() int {
	return 0
}

func (deployer *DeployerPlugin) OnApplyComponentInstanceCreate(instance *resolve.ComponentInstance) error {
	component := deployer.Next.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Instantiate component
	deployer.EventLog.WithFields(Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("[Deployer] Launching component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		codeExecutor, err := GetCodeExecutor(
			component.Code,
			instance.Key.GetKey(),
			instance.CalculatedCodeParams,
			deployer.Next.Policy.Clusters,
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

func (deployer *DeployerPlugin) OnApplyComponentInstanceUpdate(instance *resolve.ComponentInstance) error {
	component := deployer.Next.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Update component
	deployer.EventLog.WithFields(Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("[Deployer] Updating a running component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		codeExecutor, err := GetCodeExecutor(
			component.Code,
			instance.Key.GetKey(),
			instance.CalculatedCodeParams,
			deployer.Next.Policy.Clusters,
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

func (deployer *DeployerPlugin) OnApplyComponentInstanceDelete(instance *resolve.ComponentInstance) error {
	component := deployer.Prev.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]
	if component == nil {
		// This is a service instance. Do nothing
		return nil
	}

	// Delete component
	deployer.EventLog.WithFields(Fields{
		"componentKey": instance.Key,
		"component":    component.Name,
		"code":         instance.CalculatedCodeParams,
	}).Info("[Deployer] Destructing a running component instance: " + instance.Key.GetKey())

	if component.Code != nil {
		codeExecutor, err := GetCodeExecutor(
			component.Code,
			instance.Key.GetKey(),
			instance.CalculatedCodeParams,
			deployer.Prev.Policy.Clusters,
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
