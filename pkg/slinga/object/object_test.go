package object

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseObject(t *testing.T) {
	var obj Object = &BaseObject{
		"kind",
		BaseObjectMetadata{
			"uid",
			42,
			"name",
			"namespace",
		},
		map[string]string{
			"a": "b",
		},
	}

	assert.Equal(t, ObjectKind("kind"), obj.GetKind(), "Correct ObjectKind expected")
	assert.Equal(t, UID("uid"), obj.GetUID(), "Correct UID expected")
	assert.Equal(t, Generation(42), obj.GetGeneration(), "Correct Generation expected")
	assert.Equal(t, KeyFromParts("uid", 42), obj.GetKey(), "Correct Key expected")
	assert.Equal(t, "namespace", obj.GetNamespace(), "Correct Namespace expected")
	assert.Equal(t, "name", obj.GetName(), "Correct Name expected")
	assert.NotNil(t, obj.GetSpec(), "NotNil Spec expected")
}
