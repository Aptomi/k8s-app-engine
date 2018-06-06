package lang

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyDeploymentTarget(t *testing.T) {
	tc := map[string]*Target{
		"name":           {ClusterName: "name"},
		"name.suffix":    {ClusterName: "name", Suffix: "suffix"},
		"ns/name":        {ClusterNamespace: "ns", ClusterName: "name"},
		"ns/name.suffix": {ClusterNamespace: "ns", ClusterName: "name", Suffix: "suffix"},
	}

	for target, expected := range tc {
		tParsed := NewTarget(target)
		assert.Equal(t, expected.ClusterNamespace, tParsed.ClusterNamespace, "Aptomi namespace name should be correctly parsed from deployment target")
		assert.Equal(t, expected.ClusterName, tParsed.ClusterName, "Aptomi cluster name should be correctly parsed from deployment target")
		assert.Equal(t, expected.Suffix, tParsed.Suffix, "Suffix should be correctly parsed from deployment target")
	}
}
