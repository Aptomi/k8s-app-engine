package object

import (
	"fmt"
	"strconv"
	"strings"
)

// Generation represents object's "version" and starts from 0
type Generation uint64

// String returns generation as string to implement Stringer interface
func (generation Generation) String() string {
	return strconv.FormatUint(uint64(generation), 10)
}

// Next returns the next generation of the base object (current + 1)
func (generation Generation) Next() Generation {
	return generation + 1
}

// ParseGeneration returns Generation type representation of specified generation string
func ParseGeneration(gen string) Generation {
	val, err := strconv.ParseUint(gen, 10, 64)
	if err != nil {
		panic(fmt.Errorf("error while parsing generation from %s: %s", gen, err))
	}
	return Generation(val)
}

// KeySeparator used to separate parts of the Key
const KeySeparator = ":"

// Base interface represents unified object that could be stored in DB, accessed through API, etc.
type Base interface {
	GetNamespace() string
	GetKind() string
	GetName() string
	GetGeneration() Generation
	SetGeneration(Generation)
}

// GetKey returns standard key for any object.Base consists of namespace, kind and name separated by object.KeySeparator
func GetKey(obj Base) string {
	return strings.Join([]string{obj.GetNamespace(), obj.GetKind(), obj.GetName()}, KeySeparator)
}
