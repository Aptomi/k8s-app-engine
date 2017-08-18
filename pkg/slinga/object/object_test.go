package object

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testObject struct {
	Metadata

	TestField string
}

func TestMetadata(t *testing.T) {
	var obj BaseObject = &testObject{
		Metadata{
			"kind",
			"uid",
			42,
			"name",
			"namespace",
		},
		"test",
	}

	assert.Equal(t, Kind("kind"), obj.GetKind(), "Correct Kind expected")
	assert.Equal(t, UID("uid"), obj.GetUID(), "Correct UID expected")
	assert.Equal(t, Generation(42), obj.GetGeneration(), "Correct Generation expected")
	assert.Equal(t, KeyFromParts("uid", 42), obj.GetKey(), "Correct Key expected")
	assert.Equal(t, "namespace", obj.GetNamespace(), "Correct Namespace expected")
	assert.Equal(t, "name", obj.GetName(), "Correct Name expected")
}
