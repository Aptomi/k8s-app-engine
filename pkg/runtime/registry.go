package runtime

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

// Registry contains a map of objects info structures by their kind
type Registry struct {
	Kinds map[string]*Info
}

// NewRegistry creates a new Registry
func NewRegistry() *Registry {
	return &Registry{
		Kinds: make(map[string]*Info),
	}
}

// Append adds specified list of object Info into the registry
func (reg *Registry) Append(infos ...*Info) *Registry {
	for _, info := range infos {
		reg.validateInfo(info)
		reg.Kinds[info.Kind] = info
	}

	return reg
}

// Get looks up object informational structure given its kind
func (reg *Registry) Get(kind Kind) *Info {
	info, exist := reg.Kinds[kind]
	if !exist {
		panic(fmt.Sprintf("Kind '%s' isn't registered", kind))
	}

	return info
}

func (reg *Registry) validateInfo(info *Info) {
	kind := info.Kind
	if len(kind) == 0 {
		panic(fmt.Sprintf("Kind can't be empty"))
	}

	if _, exist := reg.Kinds[kind]; exist {
		panic(fmt.Sprintf("Kind can't be duplicated: %s", kind))
	}

	obj := info.New()
	if _, ok := obj.(Storable); info.Storable && !ok {
		panic(fmt.Sprintf("Kind '%s' registered as Storable but doesn't implement corresponding interface", kind))
	} else if !info.Storable && ok {
		log.Debugf("Kind '%s' registered as non-Storable but implements corresponding interface", kind)
	}
	if _, ok := obj.(Versioned); info.Versioned && !ok {
		panic(fmt.Sprintf("Kind '%s' registered as Versioned but doesn't implement corresponding interface", kind))
	} else if !info.Versioned && ok {
		log.Debugf("Kind '%s' registered as non-Versioned but implements corresponding interface", kind)
	}
}
