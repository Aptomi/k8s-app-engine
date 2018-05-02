package resolve

// DependencyResolution contains resolution status for a given dependency
type DependencyResolution struct {
	// Resolved indicates whether or not dependency has been resolved. If it has been resolved,
	// then ComponentInstanceKey will contain where it got resolved. Otherwise, you can filter event log by this
	// dependency and find out the associated events leading to an error.
	Resolved bool

	// ComponentInstanceKey holds the reference to component instance, to which dependency got resolved
	ComponentInstanceKey string
}

// Creates a new dependency resolution
func newDependencyResolution(resolveErr error, key *ComponentInstanceKey) *DependencyResolution {
	if resolveErr != nil {
		return &DependencyResolution{
			Resolved: false,
		}
	}

	return &DependencyResolution{
		Resolved:             true,
		ComponentInstanceKey: key.GetKey(),
	}
}
