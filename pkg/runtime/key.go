package runtime

import (
	"fmt"
)

// Key represents storable object key - namespace + kind + name separated by KeySeparator
type Key = string

// KeySeparator is used to separate key parts (namespace, kind, name)
const KeySeparator = "/"

// KeyFromParts returns Key build using provided parts (namespace, kind, name)
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

// KeyForStorable returns Key for storable object
func KeyForStorable(obj Storable) Key {
	return KeyFromParts(obj.GetNamespace(), obj.GetKind(), obj.GetName())
}
