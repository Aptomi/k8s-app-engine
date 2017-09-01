package actions

import "github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"

type ComponentBaseAction struct {
	// On which component instance the action is going to performed
	key string

	// Pointers to desired and actual state
	desiredState *resolve.PolicyResolution
	actualState  *resolve.PolicyResolution
}

func NewComponentBaseAction(key string, desiredState *resolve.PolicyResolution, actualState *resolve.PolicyResolution) *ComponentBaseAction {
	return &ComponentBaseAction{
		key:          key,
		desiredState: desiredState,
		actualState:  actualState,
	}
}
