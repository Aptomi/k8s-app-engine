package store

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	indexCacheMu = &sync.Mutex{}
	indexCache   = map[runtime.Kind]map[string]*Index{}
)

const LastGenIndex = ""

func Indexes(info *runtime.TypeInfo) map[string]*Index {
	indexCacheMu.Lock()
	defer indexCacheMu.Unlock()

	indexes, exist := indexCache[info.Kind]
	if !exist {
		indexes = map[string]*Index{}
		indexCache[info.Kind] = indexes

		if info.Versioned {
			indexes[LastGenIndex] = &Index{
				Scope: IndexScopeGen,
				Type:  IndexTypeLast,
				Field: LastGenIndex,
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

	if index.Field == LastGenIndex {
		return key
	}

	t := reflect.ValueOf(storable)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f := t.Field(index.fieldId)

	return index.KeyForValue(key, f.Interface(), codec)
}

func (index *Index) KeyForValue(key runtime.Key, value interface{}, codec Codec) string {
	if index.Field == LastGenIndex {
		return key
	}

	if index.Scope != IndexScopeGen {
		panic("only index scope gen is currently supported")
	}
	if index.Type != IndexTypeLast && index.Type != IndexTypeList {
		panic("only index type last or list are currently supported")
	}

	key = "@" + index.Field + "@"

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

type IndexValueList [][]byte

func (list *IndexValueList) Add(value []byte) {
	// binary search to get desired value index in the list
	valueIndex := sort.Search(len(*list), func(index int) bool {
		return bytes.Compare((*list)[index], value) >= 0
	})

	// value already present in the list
	if valueIndex < len(*list) && bytes.Equal((*list)[valueIndex], value) {
		return
	}

	// insert value into desired position
	*list = append(*list, nil)
	copy((*list)[valueIndex+1:], (*list)[valueIndex:])
	(*list)[valueIndex] = value
}

func (list *IndexValueList) Remove(value []byte) {
	// binary search to get value index in the list
	valueIndex := sort.Search(len(*list), func(index int) bool {
		return bytes.Compare((*list)[index], value) >= 0
	})

	// remove value from the list if exists
	if valueIndex < len(*list) {
		copy((*list)[valueIndex:], (*list)[valueIndex+1:])
		(*list)[len(*list)-1] = nil
		*list = (*list)[:len(*list)-1]
	}
}

func (list *IndexValueList) Contains(value []byte) bool {
	// binary search to get value index in the list
	valueIndex := sort.Search(len(*list), func(index int) bool {
		return bytes.Compare((*list)[index], value) >= 0
	})

	return valueIndex < len(*list) && bytes.Equal((*list)[valueIndex], value)
}
