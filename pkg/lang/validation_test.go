package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyObjectValidationIdentifiers(t *testing.T) {
	passed := []object.Base{
		makeService("test"),
		makeService("good_name"),
		makeService("excellent-name_239"),
	}
	failed := []object.Base{
		makeService("_test"),
		makeService("12-good_name"),
		makeService("excellent-n#ame_239"),
		makeService("excellent-n$ame_239"),
	}
	runValidationTests(t, passed, failed)
}

func TestPolicyObjectValidationDuplicateIdentifiers(t *testing.T) {
	// TODO: implement for duplicate identifiers
	passed := []object.Base{}
	failed := []object.Base{}
	runValidationTests(t, passed, failed)
}

func TestPolicyObjectValidationExpressions(t *testing.T) {
	passed := []object.Base{
		makeRule(100, "specialname + specialvalue == 'b'", "true", "false"),
	}
	failed := []object.Base{
		makeRule(100, "specialname + '123')(((", "true", "false"),
		makeRule(100, "true", "512 + 813 >>><<< 'a'", "false"),
		makeRule(100, "false", "true", "a + b + c__&"),
	}
	runValidationTests(t, passed, failed)
}

func TestPolicyObjectValidationRuleWeight(t *testing.T) {
	passed := []object.Base{
		makeRule(1, "true", "true", "true"),
	}
	failed := []object.Base{
		makeRule(-1, "true", "true", "true"),
	}
	runValidationTests(t, passed, failed)
}

func TestPolicyObjectValidationClusterTypes(t *testing.T) {
	passed := []object.Base{
		makeCluster("kubernetes"),
	}
	failed := []object.Base{
		makeCluster("unknown"),
	}
	runValidationTests(t, passed, failed)
}

func runValidationTests(t *testing.T, passed []object.Base, failed []object.Base) {
	policy := NewPolicy()
	for _, obj := range passed {
		err := policy.AddObject(obj)
		assert.NoError(t, err, "Policy.AddObject() call should be successful (object should pass validation)")
	}
	for _, obj := range failed {
		err := policy.AddObject(obj)
		assert.Error(t, err, "Policy.AddObject() call should return an error (object should fail validation)")
	}
}

func makeRule(weight int, all string, any string, none string) *Rule {
	return &Rule{
		Metadata: Metadata{
			Kind:      RuleObject.Kind,
			Namespace: "main",
			Name:      "rule",
		},
		Weight: weight,
		Criteria: &Criteria{
			RequireAll:  []string{all},
			RequireAny:  []string{any},
			RequireNone: []string{none},
		},
	}
}

func makeCluster(clusterType string) *Cluster {
	return &Cluster{
		Metadata: Metadata{
			Kind:      ClusterObject.Kind,
			Namespace: object.SystemNS,
			Name:      "cluster",
		},
		Type: clusterType,
	}
}

func makeService(name string) *Service {
	return &Service{
		Metadata: Metadata{
			Kind:      ServiceObject.Kind,
			Namespace: "main",
			Name:      name,
		},
		Owner: "1",
	}
}
