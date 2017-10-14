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

func ParseGeneration(gen string) Generation {
	val, err := strconv.ParseUint(gen, 10, 64)
	if err != nil {
		panic(fmt.Errorf("error while parsing generation from %s: %s", gen, err))
	}
	return Generation(val)
}

// KeySeparator used to separate parts of the Key
const KeySeparator = ":"

/*
// Key represents human-readable unified object's key that can always identify any object in Aptomi.
// It consists of several parts - [<domain>:]<namespace>:<kind>:<name>:<rand_addon>:<generation>, where:
// * domain - Aptomi deployment name, optional
// * namespace
// * kind
// * name
// * rand_addon - random 6 letters added to the Key to be able to differentiate objects re-created with the same name, unique for any specific namespace:kind:name
// * generation - object "version", starts from 1
// So, it means that Key could be used to reference concrete object with concrete generation.
type Key string

type KeyParts struct {
	Domain     string
	Namespace  string
	Kind       string
	Name       string
	RandAddon  string
	Generation Generation
}

func (key Key) Parts() (*KeyParts, error) {
	parts := strings.Split(string(key), KeySeparator)
	partsLen := len(parts)

	domain := ""

	// todo(slukjanov): support non-namespaced objects like clusters? userproviders? etc

	if partsLen == 6 {
		domain = parts[0]
		parts = parts[1:]
	} else if partsLen != 5 {
		return nil, fmt.Errorf("Can't parse key: %s", key)
	}

	gen, err := strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Can't parse generation of key %s with error: %s", key, err)
	}

	return &KeyParts{
		Domain:     domain,
		Namespace:  parts[0],
		Kind:       parts[1],
		Name:       parts[2],
		RandAddon:  parts[3],
		Generation: Generation(gen),
	}, nil
}

// KeyFromParts return uid and generation combined into Key
func KeyFromParts(domain string, namespace string, kind string, name string, randAddon string, generation Generation) Key {
	if len(domain) > 0 {
		return Key(fmt.Sprintf("%s:%s:%s:%s:%s:%s", domain, namespace, kind, name, randAddon, generation))
	}
	return Key(fmt.Sprintf("%s:%s:%s:%s:%s", namespace, kind, name, randAddon, generation))
}
*/

// Base interface represents unified object that could be stored in DB, accessed through API, etc.
type Base interface {
	GetNamespace() string
	GetKind() string
	GetName() string
	GetGeneration() Generation
	SetGeneration(Generation)
}

func GetKey(obj Base) string {
	return strings.Join([]string{obj.GetNamespace(), obj.GetKind(), obj.GetName()}, KeySeparator)
}
