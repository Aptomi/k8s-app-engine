package runtime

import (
	"fmt"
)

// TypeInfo represents list of additional characteristics of the runtime object
type TypeInfo struct {
	Kind        Kind
	Storable    bool
	Versioned   bool
	Deletable   bool
	Constructor Constructor
}

// Constructor is a function to get instance of the specific object
type Constructor func() Object

// New creates a new instance of the specific object defined in TypeInfo
func (info *TypeInfo) New() Object {
	return info.Constructor()
}

// GetTypeKind returns TypeKind instance for the object described by info
func (info *TypeInfo) GetTypeKind() TypeKind {
	return TypeKind{Kind: info.Kind}
}

// Types contains a map of objects info structures by their kind
type Types struct {
	Kinds map[string]*TypeInfo
}

// NewTypes creates a new Types
func NewTypes() *Types {
	return &Types{
		Kinds: make(map[string]*TypeInfo),
	}
}

// AppendAllTypes concatenates all provided info slices into a single info slice
func AppendAllTypes(all ...[]*TypeInfo) []*TypeInfo {
	result := make([]*TypeInfo, 0)

	for _, infos := range all {
		result = append(result, infos...)
	}

	return result
}

// Append adds specified list of object TypeInfo into the registry
func (reg *Types) Append(infos ...*TypeInfo) *Types {
	for _, info := range infos {
		reg.validateInfo(info)
		reg.Kinds[info.Kind] = info
	}

	return reg
}

// Get looks up object informational structure given its kind
func (reg *Types) Get(kind Kind) *TypeInfo {
	info, exist := reg.Kinds[kind]
	if !exist {
		panic(fmt.Sprintf("Kind '%s' isn't registered", kind))
	}

	return info
}

func (reg *Types) validateInfo(info *TypeInfo) {
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
	} /* else if !info.Storable && ok {
		log.Debugf("Kind '%s' registered as non-Storable but implements corresponding interface", kind)
	} */
	if _, ok := obj.(Versioned); info.Versioned && !ok {
		panic(fmt.Sprintf("Kind '%s' registered as Versioned but doesn't implement corresponding interface", kind))
	} /* else if !info.Versioned && ok {
		log.Debugf("Kind '%s' registered as non-Versioned but implements corresponding interface", kind)
	} */
}
