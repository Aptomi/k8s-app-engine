package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/deployment"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	log "github.com/Sirupsen/logrus"
)

type EngineApply struct {
	// Diff to be applied
	diff *diff.ServiceUsageStateDiff

	// Progress indicator
	progress progress.ProgressIndicator
}

func NewEngineApply(diff *diff.ServiceUsageStateDiff) *EngineApply {
	return &EngineApply{
		diff:     diff,
		progress: progress.NewProgressConsole(),
	}
}

// Apply method applies all changes via executors, saves usage state in Aptomi DB
func (apply *EngineApply) Apply() error {
	// initialize progress indicator
	apply.progress.SetTotal(apply.diff.GetApplyProgressLength())

	// process all actions
	err := apply.processDestructions()
	if err != nil {
		return fmt.Errorf("Error while destructing components")
	}
	err = apply.processUpdates()
	if err != nil {
		return fmt.Errorf("Error while updating components")
	}
	err = apply.processInstantiations()
	if err != nil {
		return fmt.Errorf("Error while instantiating components")
	}

	// call plugins to perform their actions
	// TODO: error handling from plugins
	for _, plugin := range apply.diff.Plugins {
		plugin.Apply(apply.progress)
	}

	// Finalize progress indicator
	apply.progress.Done()
	return nil
}

func (apply *EngineApply) processInstantiations() error {
	// Process instantiations in the right order
	for _, key := range apply.diff.Next.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be instantiated?
		if _, ok := apply.diff.ComponentInstantiate[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Create")

			instance := apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key]
			component := apply.diff.Next.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]

			if component == nil {
				Debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    instance.Key.ServiceName,
				}).Info("Instantiating service")

				// TODO: add processing code
			} else {
				Debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Instantiating component")

				if component.Code != nil {
					codeExecutor, err := deployment.GetCodeExecutor(
						component.Code,
						key,
						apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key].CalculatedCodeParams,
						apply.diff.Next.Policy.Clusters,
					)
					if err != nil {
						return err
					}

					err = codeExecutor.Install()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (apply *EngineApply) processUpdates() error {
	// Process updates in the right order
	for _, key := range apply.diff.Next.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be updated?
		if _, ok := apply.diff.ComponentUpdate[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Update")

			instance := apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key]
			component := apply.diff.Prev.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]
			if component == nil {
				Debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    instance.Key.ServiceName,
				}).Info("Updating service")

				// TODO: add processing code
			} else {
				Debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Updating component")

				if component.Code != nil {
					codeExecutor, err := deployment.GetCodeExecutor(
						component.Code,
						key,
						apply.diff.Next.State.ResolvedData.ComponentInstanceMap[key].CalculatedCodeParams,
						apply.diff.Next.Policy.Clusters,
					)
					if err != nil {
						return err
					}
					err = codeExecutor.Update()
					if err != nil {
						return err
					}

				}
			}
		}
	}
	return nil
}

func (apply *EngineApply) processDestructions() error {
	// Process destructions in the right order
	for _, key := range apply.diff.Prev.State.ResolvedData.ComponentProcessingOrder {
		// Does it need to be destructed?
		if _, ok := apply.diff.ComponentDestruct[key]; ok {
			// Advance progress indicator
			apply.progress.Advance("Delete")

			instance := apply.diff.Prev.State.ResolvedData.ComponentInstanceMap[key]
			component := apply.diff.Prev.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]
			if component == nil {
				Debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    instance.Key.ServiceName,
				}).Info("Destructing service")

				// TODO: add processing code
			} else {
				Debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Destructing component")

				if component.Code != nil {
					codeExecutor, err := deployment.GetCodeExecutor(
						component.Code,
						key,
						apply.diff.Prev.State.ResolvedData.ComponentInstanceMap[key].CalculatedCodeParams,
						apply.diff.Prev.Policy.Clusters,
					)
					if err != nil {
						return err
					}
					err = codeExecutor.Destroy()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
