package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"time"
)

func updateActualStateFromDesired(componentKey string, context *action.Context, createNow bool, updateNow bool, createIfNotExists bool) error {
	// get instance from actual state
	instanceActual := context.ActualState.ComponentInstanceMap[componentKey]

	// if it doesn't exist and we are not forced to create it, then return
	if instanceActual == nil && !createIfNotExists {
		return nil
	}

	// preserve previous CreatedAt/UpdatedAt dates before overwriting
	timeCreated := time.Now()
	timeUpdated := time.Now()
	if instanceActual != nil {
		if !createNow {
			timeCreated = instanceActual.CreatedAt
		}
		if !updateNow {
			timeUpdated = instanceActual.UpdatedAt
		}
	}

	// get instance from desired state
	instance := context.DesiredState.ComponentInstanceMap[componentKey]
	if instance == nil {
		panic(fmt.Sprintf("component instance not found in desired state: %s", componentKey))
	}

	// modify create/update times, copy it over to the actual state
	instance.UpdateTimes(timeCreated, timeUpdated)
	context.ActualState.ComponentInstanceMap[componentKey] = instance

	// save actual state
	err := context.ActualStateUpdater.Save(instance)
	if err != nil {
		return fmt.Errorf("error while updating actual state: %s", err)
	}
	return nil
}

func updateComponentInActualState(componentKey string, context *action.Context) error {
	instance := context.ActualState.ComponentInstanceMap[componentKey]
	err := context.ActualStateUpdater.Save(instance)
	if err != nil {
		return fmt.Errorf("error while updating actual state: %s", err)
	}
	return nil
}

func deleteComponentFromActualState(componentKey string, context *action.Context) error {
	// delete component from the actual state
	delete(context.ActualState.ComponentInstanceMap, componentKey)
	err := context.ActualStateUpdater.Delete(resolve.KeyForComponentKey(componentKey))
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}
