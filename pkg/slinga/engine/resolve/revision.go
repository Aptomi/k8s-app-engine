package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
)

// Revision represents an aggregated object after policy resolution (input+external+resolved objects)
type Revision struct {
	// Input
	Policy *language.PolicyNamespace

	// Output
	Resolution *PolicyResolution

	// External data
	UserLoader language.UserLoader `yaml:"-"`
}

// NewRevision creates a new revision
func NewRevision(policy *language.PolicyNamespace, resolution *PolicyResolution, userLoader language.UserLoader) *Revision {
	return &Revision{
		Policy:     policy,
		Resolution: resolution,
		UserLoader: userLoader,
	}
}

func (revision *Revision) Save() {
	fileName := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypePolicyResolution, "db.yaml")
	yaml.SaveObjectToFile(fileName, revision)
}

// LoadRevision loads revision from a file under Aptomi DB
func LoadRevision() *Revision {
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), lastRevision.GetRunDirectory(), TypePolicyResolution, "db.yaml")
	result := loadRevisionFromFile(fileName)
	if result.Policy == nil {
		result.Policy = language.NewPolicyNamespace()
	}
	if result.Resolution == nil {
		result.Resolution = NewPolicyResolution()
	}
	if result.UserLoader == nil {
		result.UserLoader = language.NewAptomiUserLoader()
	}
	return result
}

// LoadRevisionsAll loads all revisions from files under Aptomi DB
func LoadRevisionsAll() map[int]*Revision {
	result := make(map[int]*Revision)
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	for rev := lastRevision; rev > LastRevisionAbsentValue; rev-- {
		fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), rev.GetRunDirectory(), TypePolicyResolution, "db.yaml")
		revision := loadRevisionFromFile(fileName)
		result[int(rev)] = revision
	}
	return result
}

// Loads revision from file
func loadRevisionFromFile(fileName string) *Revision {
	return yaml.LoadObjectFromFileDefaultEmpty(fileName, new(Revision)).(*Revision)
}
