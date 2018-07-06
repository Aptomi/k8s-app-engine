package registry

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

func (reg *defaultRegistry) GetActualState() (*resolve.PolicyResolution, error) {
	var instances []*resolve.ComponentInstance
	// todo we should support getting all objects by kind?
	err := reg.store.Find(resolve.TypeComponentInstance.Kind, &instances, store.WithKeyPrefix(runtime.SystemNS+"/"+resolve.TypeComponentInstance.Kind))
	if err != nil {
		return nil, fmt.Errorf("error while getting all component instances: %s", err)
	}

	actualState := resolve.NewPolicyResolution()
	for _, instance := range instances {
		key := instance.GetKey()
		actualState.ComponentInstanceMap[key] = instance
	}

	return actualState, nil
}
