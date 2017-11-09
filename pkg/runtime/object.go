package runtime

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

type Kind = string

var Unknown = Kind("unknown")

type Object interface {
	GetKind() Kind
}

type Storable interface {
	Object
	GetName() string
	GetNamespace() string
}

type Versioned interface {
	Storable
	GetGeneration() Generation
	SetGeneration(gen Generation)
}

type Info struct {
	Kind        Kind
	Storable    bool
	Versioned   bool
	Constructor Constructor
}

// Constructor is a function to get instance of the specific object
type Constructor func() Object

// New creates a new instance of the specific object defined in Info
func (info *Info) New() Object {
	return info.Constructor()
}

func (info *Info) GetTypeKind() TypeKind {
	return TypeKind{info.Kind}
}

type Key = string

const KeySeparator = "/"

func KeyFromParts(namespace string, kind Kind, name string) Key {
	if len(namespace) == 0 {
		panic(fmt.Sprintf("Key couldn't be created with empty namespace"))
	}
	if len(kind) == 0 {
		panic(fmt.Sprintf("Key couldn't be created with empty kind"))
	}

	key := namespace + KeySeparator + kind

	if len(name) > 0 {
		key += KeySeparator + name
	}

	return key
}

func KeyFromStorable(obj Storable) Key {
	return KeyFromParts(obj.GetNamespace(), obj.GetKind(), obj.GetName())
}

type TypeKind struct {
	Kind Kind
}

func (tk *TypeKind) GetKind() Kind {
	return tk.Kind
}

type GenerationMetadata struct {
	Generation Generation
}
