package resolve

import (
	"strings"
	"testing"

	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/builder"
	"github.com/stretchr/testify/assert"
)

func TestComponentKeyCopy(t *testing.T) {
	{
		key := makeKey(false)
		keyCopy := key.MakeCopy()
		assert.Equal(t, key.GetKey(), keyCopy.GetKey(), "Component key should be copied successfully")
	}

	{
		key := makeKey(true)
		keyCopy := key.MakeCopy()
		assert.Equal(t, key.GetKey(), keyCopy.GetKey(), "Bundle key (root) should be copied successfully")
	}
}

func TestComponentKeyParent(t *testing.T) {
	{
		key := makeKey(false)
		keyParent := key.GetParentBundleKey()
		k1 := strings.Split(key.GetKey(), componentInstanceKeySeparator)
		k2 := strings.Split(keyParent.GetKey(), componentInstanceKeySeparator)
		assert.Equal(t, len(k1), len(k2), "Component key and its parent should have the same number of parts")
		for i := 0; i < len(k1)-1; i++ {
			assert.Equal(t, k1[i], k2[i], "Parent for component key should have the same parts, except point to a bundle")
		}
		assert.Equal(t, componentRootName, k2[len(k2)-1], "Parent for component key should have the same parts, except point to a bundle")
	}

	{
		key := makeKey(true)
		keyParent := key.GetParentBundleKey()
		assert.Equal(t, key.GetKey(), keyParent.GetKey(), "Parent for bundle key (root) should point to itself")
	}
}

func TestComponentKeyUnsafe(t *testing.T) {
	key := makeKeyUnsafe()
	k := strings.Split(key.GetKey(), componentInstanceKeySeparator)
	expected := []string{componentUnresolvedName, componentUnresolvedName, componentUnresolvedName, componentUnresolvedName, componentUnresolvedName, componentUnresolvedName, componentRootName}
	assert.Equal(t, len(k), len(expected), "When policy objects are nil, component key with correct number of entries should still be generated: %s", key.GetKey())
	for i := range expected {
		assert.Equal(t, expected[i], k[i], "When policy objects are nil, component key should still be generated: %s", key.GetKey())
	}
}

func makeKey(root bool) *ComponentInstanceKey {
	b := builder.NewPolicyBuilder()
	bundle := b.AddBundle()

	var component *lang.BundleComponent
	if !root {
		component = b.AddBundleComponent(bundle, b.CodeComponent(nil, nil))
	}

	service := b.AddService(bundle, b.CriteriaTrue())
	key := NewComponentInstanceKey(
		b.AddCluster(),
		"suffix",
		service,
		service.Contexts[0],
		[]string{"x", "y", "z"},
		bundle,
		component,
	)
	return key
}

func makeKeyUnsafe() *ComponentInstanceKey {
	return NewComponentInstanceKey(
		nil,
		"",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}
