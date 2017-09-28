package resolve

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponentKeyCopy(t *testing.T) {
	policyNS := loadUnitTestsPolicy().Namespace["main"]

	key := NewComponentInstanceKey(
		policyNS.Clusters["cluster-us-west"],
		policyNS.Contracts["zookeeper"],
		policyNS.Contracts["zookeeper"].Contexts[0],
		[]string{"x", "y", "z"},
		policyNS.Services["zookeeper"],
		policyNS.Services["zookeeper"].Components[0],
	)

	keyCopy := key.MakeCopy()

	assert.Equal(t, key.GetKey(), keyCopy.GetKey(), "Component key should be copied successfully")
}
