package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
)

// ResolvedState represents an aggregated object after policy resolution (input+external+resolved objects)
type ResolvedState struct {
	// Input
	Policy *language.PolicyNamespace

	// Output
	State *ServiceUsageState

	// External data
	UserLoader language.UserLoader
}

// NewResolvedState creates a new revision state
func NewResolvedState(policy *language.PolicyNamespace, state *ServiceUsageState, userLoader language.UserLoader) *ResolvedState {
	return &ResolvedState{
		Policy:     policy,
		State:      state,
		UserLoader: userLoader,
	}
}

func (resolvedState *ResolvedState) Save() {
	// Save usage state
	fileName := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypePolicyResolution, "db.yaml")
	yaml.SaveObjectToFile(fileName, resolvedState)
}

// LoadServiceUsageState loads usage state from a file under Aptomi DB
func LoadResolvedState() *ResolvedState {
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), lastRevision.GetRunDirectory(), TypePolicyResolution, "db.yaml")
	result := loadResolvedStateFromFile(fileName)
	if result.Policy == nil {
		result.Policy = language.NewPolicyNamespace()
	}
	if result.State == nil {
		result.State = NewServiceUsageState()
	}
	if result.UserLoader == nil {
		result.UserLoader = language.NewAptomiUserLoader()
	}
	return result
}

// LoadServiceUsageStatesAll loads all usage states from files under Aptomi DB
func LoadResolvedStatesAll() map[int]*ResolvedState {
	result := make(map[int]*ResolvedState)
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	for rev := lastRevision; rev > LastRevisionAbsentValue; rev-- {
		fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), rev.GetRunDirectory(), TypePolicyResolution, "db.yaml")
		state := loadResolvedStateFromFile(fileName)
		result[int(rev)] = state
	}
	return result
}

// Loads usage state from file
func loadResolvedStateFromFile(fileName string) *ResolvedState {
	return yaml.LoadObjectFromFileDefaultEmpty(fileName, new(ResolvedState)).(*ResolvedState)
}
