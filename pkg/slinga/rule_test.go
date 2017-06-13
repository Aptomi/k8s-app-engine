package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadRules(t *testing.T) {
	rules := LoadRulesFromDir("testdata/unittests")

	assert.Equal(t, 1, len(rules.Rules), "Correct number of rules should be loaded")
	assert.Equal(t, "compromised", rules.Rules["ingress"][0].FilterServices.Cluster.Accept[0])
	assert.Equal(t, "sensetive", rules.Rules["ingress"][0].FilterServices.Labels.Accept[0])
}
