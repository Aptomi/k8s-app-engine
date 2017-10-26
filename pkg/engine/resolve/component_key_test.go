package resolve

import (
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponentKeyCopy(t *testing.T) {
	// create component key
	b := builder.NewPolicyBuilder()
	service := b.AddService()
	component := b.AddServiceComponent(service, b.CodeComponent(nil, nil))
	contract := b.AddContract(service, b.CriteriaTrue())
	key := NewComponentInstanceKey(
		b.AddCluster(),
		contract,
		contract.Contexts[0],
		[]string{"x", "y", "z"},
		service,
		component,
	)

	// make component key copy
	keyCopy := key.MakeCopy()

	// check that both keys as strings are identical
	assert.Equal(t, key.GetKey(), keyCopy.GetKey(), "Component key should be copied successfully")
}
