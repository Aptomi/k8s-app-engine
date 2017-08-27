package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadRules(t *testing.T) {
	policy := LoadUnitTestsPolicy()
	assert.Equal(t, 2, len(policy.Rules.Rules), "Correct number of rule action types should be loaded")
	assert.Equal(t, "compromised", policy.Rules.Rules["ingress"][0].FilterServices.Cluster.RequireAny[0])
	assert.Equal(t, "sensitive", policy.Rules.Rules["ingress"][0].FilterServices.Labels.RequireAny[0])
}
