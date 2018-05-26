package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"time"
)

func createComponentInActualState(componentKey string, context *action.Context) error {
	// get instance from desired state
	instance := context.DesiredState.ComponentInstanceMap[componentKey]
	if instance == nil {
		panic(fmt.Sprintf("component instance not found in desired state: %s", componentKey))
	}

	// update timestamps
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// move it over to the actual state
	context.ActualState.ComponentInstanceMap[componentKey] = instance

	// save component instance in the actual state store
	err := context.ActualStateUpdater.Save(instance)
	if err != nil {
		return fmt.Errorf("error while updating actual state: %s", err)
	}
	return nil
}

func updateComponentInActualState(componentKey string, context *action.Context) error {
	// look up an existing component in the actual state
	instance := context.ActualState.ComponentInstanceMap[componentKey]

	// update timestamp
	instance.UpdatedAt = time.Now()

	// save component instance in the actual state store
	err := context.ActualStateUpdater.Save(instance)
	if err != nil {
		return fmt.Errorf("error while updating actual state: %s", err)
	}
	return nil
}

func deleteComponentFromActualState(componentKey string, context *action.Context) error {
	// delete an existing component from the actual state map
	delete(context.ActualState.ComponentInstanceMap, componentKey)

	// delete an existing component from the actual state store
	err := context.ActualStateUpdater.Delete(resolve.KeyForComponentKey(componentKey))
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}
