package resolve

// ClaimResolution contains resolution status for a given claim
type ClaimResolution struct {
	// Resolved indicates whether or not claim has been resolved. If it has been resolved,
	// then ComponentInstanceKey will contain where it got resolved. Otherwise, you can filter event log by this
	// claim and find out the associated events leading to an error.
	Resolved bool

	// ComponentInstanceKey holds the reference to component instance, to which claim got resolved
	ComponentInstanceKey string
}

// Creates a new claim resolution
func newClaimResolution(resolved bool, key string) *ClaimResolution {
	return &ClaimResolution{
		Resolved:             resolved,
		ComponentInstanceKey: key,
	}
}
