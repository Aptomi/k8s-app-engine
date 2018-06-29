package lang

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectsInstantiate(t *testing.T) {
	for _, obj := range PolicyTypes {
		objInstance := obj.New()
		structName := reflect.TypeOf(objInstance).Elem().Name()
		assert.Contains(t, obj.Kind, strings.ToLower(structName), "%s instantiated to %s", structName, obj.Kind)
	}
}
