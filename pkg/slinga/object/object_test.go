package object

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testObject struct {
	Metadata

	TestField string
}

var testObjectInstance = &testObject{
	Metadata{
		"test",
		"4d0e5391-83ef-11e7-876b-784f435b826b",
		42,
		"name",
		"namespace",
	},
	"test",
}

func assertTestObjectInstance(t *testing.T, obj BaseObject) {
	assert.Equal(t, Kind("test"), obj.GetKind(), "Correct Kind expected")
	assert.Equal(t, UID("4d0e5391-83ef-11e7-876b-784f435b826b"), obj.GetUID(), "Correct UID expected")
	assert.Equal(t, Generation(42), obj.GetGeneration(), "Correct Generation expected")
	assert.Equal(t, KeyFromParts("4d0e5391-83ef-11e7-876b-784f435b826b", 42), obj.GetKey(), "Correct Key expected")
	assert.Equal(t, "namespace", obj.GetNamespace(), "Correct Namespace expected")
	assert.Equal(t, "name", obj.GetName(), "Correct Name expected")
}

func TestMetadata(t *testing.T) {
	assertTestObjectInstance(t, testObjectInstance)
}
