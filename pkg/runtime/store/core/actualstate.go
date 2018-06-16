package core

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

func (ds *defaultStore) GetActualState() (*resolve.PolicyResolution, error) {
	actualState := resolve.NewPolicyResolution()

	instances, err := ds.store.List(runtime.KeyFromParts(runtime.SystemNS, resolve.ComponentInstanceObject.Kind, ""))
	if err != nil {
		return nil, fmt.Errorf("error while getting all component instances: %s", err)
	}

	for _, instanceObj := range instances {
		if instance, ok := instanceObj.(*resolve.ComponentInstance); ok {
			key := instance.GetKey()
			actualState.ComponentInstanceMap[key] = instance
		}
	}

	return actualState, nil
}
