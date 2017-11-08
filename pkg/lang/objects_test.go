package lang

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestObjectsInstantiate(t *testing.T) {
	for _, obj := range PolicyObjects {
		objInstance := obj.New()
		structName := reflect.TypeOf(objInstance).Elem().Name()
		assert.Contains(t, obj.Kind, strings.ToLower(structName), "%s instantiated to %s", structName, obj.Kind)
	}
}
