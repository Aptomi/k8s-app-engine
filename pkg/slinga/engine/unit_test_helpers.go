package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
)

func loadUnitTestsPolicy() *Policy {
	policyLoader := NewSlingaObjectDatabaseDir("../testdata/unittests")
	policy := policyLoader.LoadPolicyObjects(-1, "")
	return policy
}

func emulateSaveAndLoadState(state ServiceUsageState) ServiceUsageState {
	// Emulate saving and loading again
	savedObjectAsString := yaml.SerializeObject(state)
	userLoader := NewUserLoaderFromDir("../testdata/unittests")
	loadedObject := ServiceUsageState{userLoader: userLoader}
	yaml.DeserializeObject(savedObjectAsString, &loadedObject)
	return loadedObject
}
