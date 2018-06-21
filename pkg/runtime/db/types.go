package db

import (
	"fmt"
	"sync"
)

type Kind = string
type Key = string

type Storable interface {
	GetKind() Kind
	GetKey() Key
}

type Versioned interface {
	Storable
	GetVersion()
	SetVersion()
}

type TypeInfo interface {
	Kind() Kind
	Indexes() map[string]Index
	IsVersioned() bool
	//New() Storable
	//NewSlice() []Storable
}

var (
	typeInfosMu sync.RWMutex
	typeInfos   = make(map[Kind]TypeInfo)
)

func RegisterType(storable Storable, indexes ...*Index) {
	typeInfosMu.Lock()
	defer typeInfosMu.Unlock()

	if storable == nil {
		panic("db: can't register nil type")
	}

	kind := storable.GetKind()
	if len(kind) == 0 {
		panic(fmt.Sprintf("db: can't register type with empty kind: %T", storable))
	}

	if _, duplicated := typeInfos[kind]; duplicated {
		panic(fmt.Sprintf("db: register called twice for type: %T", storable))
	}

	typeInfos[kind] = buildTypeInfo(storable, indexes)
}

func buildTypeInfo(storable Storable, indexes []*Index) TypeInfo {
	// todo verify object, indexes and fill in the type info object
	return &reflectTypeInfo{}
}

type reflectTypeInfo struct {
	kind        Kind
	indexes     map[string]Index
	isVersioned bool
}

func (info *reflectTypeInfo) Kind() Kind {
	return info.kind
}

func (info *reflectTypeInfo) Indexes() map[string]Index {
	return info.indexes
}

func (info *reflectTypeInfo) IsVersioned() bool {
	return info.isVersioned
}
