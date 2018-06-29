package newdb

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/runtime"
)

// list of participating values for specific index
//type Index func(name string, value Storable) ([]interface{}, error)

type IndexType int

const (
	IndexTypeUndef = iota
	IndexTypeLast
	IndexTypeList
)

func (indexType IndexType) String() string {
	indexTypes := [...]string{
		"last",
		"list",
	}

	if indexType < 1 || indexType > 2 {
		panic(fmt.Sprintf("unknown index type: %d", indexType))
	}

	return indexTypes[indexType-1]
}

type IndexScope int

const (
	IndexScopeUndef = iota
	IndexScopeGen
	IndexScopeKey
)

func (indexScope IndexScope) String() string {
	indexScopes := [...]string{
		"gen",
		"key",
	}

	if indexScope < 1 || indexScope > 2 {
		panic(fmt.Sprintf("unknown index scope: %d", indexScope))
	}

	return indexScopes[indexScope-1]
}

type IndexValue func(storable runtime.Storable) []byte

type Index struct {
	Field string
	Type  IndexType
	Scope IndexScope
}

func (index *Index) Key(objectKey runtime.Key, fieldValue interface{}) string {
	key := index.Scope.String() + "/"

	if len(objectKey) == 0 {
		key += objectKey + "/"
	}

	key += index.Type.String()

	// todo not use Sprintf?
	if fieldValue != nil {
		key += fmt.Sprintf("%s", fieldValue)
	}

	return key
}

var LastGenIndex = &Index{
	Field: "",
	Type:  IndexTypeLast,
	Scope: IndexScopeGen,
}

var RevisionGenForPolicyGenIndex = &Index{
	Field: "PolicyGen",
	Type:  IndexTypeList,
	Scope: IndexScopeGen,
}

/*

Indexes examples:

* /objects/<kind>/<object_key>@<version> - object itself
* /indexes/<kind>/<index_name>@<index_key> = <object_key>
	<index_key> - some value compiled from the object
	last version index will have func that returns <object_key>
	<index_type> = last

*/
