package object

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type T1 struct {
	Metadata

	T1F string
}

type T2 struct {
	Metadata

	T2F string
}

var (
	TestObjects = []BaseObject{
		&T1{
			Metadata{
				"t1",
				"uid-1",
				11,
				"name-1",
				"namespace-1",
			},
			"t1f-1",
		},
		&T1{
			Metadata{
				"t1",
				"uid-2",
				12,
				"name-2",
				"namespace-2",
			},
			"t1f-2",
		},
		&T2{
			Metadata{
				"t2",
				"uid-3",
				21,
				"name-3",
				"namespace-3",
			},
			"t2f-1",
		},
		&T2{
			Metadata{
				"t2",
				"uid-4",
				22,
				"name-4",
				"namespace-4",
			},
			"t2f-2",
		},
	}
)

func TestMetadata(t *testing.T) {
	obj := TestObjects[0]

	assert.Equal(t, Kind("t1"), obj.GetKind(), "Correct Kind expected")
	assert.Equal(t, UID("uid-1"), obj.GetUID(), "Correct UID expected")
	assert.Equal(t, Generation(11), obj.GetGeneration(), "Correct Generation expected")
	assert.Equal(t, KeyFromParts("uid-1", 11), obj.GetKey(), "Correct Key expected")
	assert.Equal(t, "namespace-1", obj.GetNamespace(), "Correct Namespace expected")
	assert.Equal(t, "name-1", obj.GetName(), "Correct Name expected")
}
