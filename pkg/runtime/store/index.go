package store

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	indexCacheMu = &sync.Mutex{}
	indexCache   = map[runtime.Kind]map[string]*Index{}
)

func Indexes(info *runtime.TypeInfo) map[string]*Index {
	indexCacheMu.Lock()
	defer indexCacheMu.Unlock()

	indexes, exist := indexCache[info.Kind]
	if !exist {
		indexes = map[string]*Index{}

		if info.Versioned {
			indexes[""] = &Index{
				Scope: IndexScopeGen,
				Type:  IndexTypeLast,
				Field: "",
			}
		}

		t := reflect.TypeOf(info.New())
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("store")

			// todo better tag, probably `index:"gen,list"`
			if strings.Contains(tag, "gen_index") {
				// todo validate that field is accessible
				// todo cache reflection objects

				indexes[f.Name] = &Index{
					Scope:   IndexScopeGen,
					Type:    IndexTypeList,
					Field:   f.Name,
					fieldId: i,
				}
			}
		}
	}

	return indexes
}

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
	Scope   IndexScope
	Type    IndexType
	Field   string
	fieldId int
}

func (index *Index) KeyForStorable(storable runtime.Storable, codec Codec) string {
	key := runtime.KeyForStorable(storable)

	if index.Field == "" {
		return key
	}

	key += "@" + index.Field + "@"

	t := reflect.ValueOf(storable)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f := t.Field(index.fieldId)

	// treat string separately
	if f.Kind() == reflect.String {
		return key + f.String()
	}

	//fmt.Println("fieldType", f.Type())

	// treat generation as a special case
	if reflect.TypeOf(runtime.FirstGen) == f.Type() {
		return key + strconv.FormatUint(f.Uint(), 10)
	}

	data, err := codec.Marshal(f.Interface())
	if err != nil {
		panic(fmt.Sprintf("error marshalling index value %s.%s=%v", storable.GetKind(), index.Field, f.Interface()))
	}

	return key + string(data)
}

func (index *Index) KeyForValue(key runtime.Key, value interface{}, codec Codec) string {
	if index.Field == "" {
		return key
	}

	key += "@" + index.Field + "@"

	if valueStr, ok := value.(string); ok {
		return key + valueStr
	}

	if valueGen, ok := value.(runtime.Generation); ok {
		return key + valueGen.String()
	}

	data, err := codec.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("error marshalling index value %s=%v", index.Field, value))
	}

	return key + string(data)
}
