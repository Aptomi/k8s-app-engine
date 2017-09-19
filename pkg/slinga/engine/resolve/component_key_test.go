package resolve

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponentKeyCopy(t *testing.T) {
	policy := loadUnitTestsPolicy()

	key := NewComponentInstanceKey(
		policy.Clusters["cluster-us-west"],
		policy.Contracts["zookeeper"],
		policy.Contracts["zookeeper"].Contexts[0],
		[]string{"x", "y", "z"},
		policy.Services["zookeeper"],
		policy.Services["zookeeper"].Components[0],
	)

	keyCopy := key.MakeCopy()

	assert.Equal(t, key.GetKey(), keyCopy.GetKey(), "Component key should be copied successfully")
}
