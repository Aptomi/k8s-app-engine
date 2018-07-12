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
	indexCache   = map[runtime.Kind]*Indexes{}
)

// LastGenIndex is the name of the index to get last generation for an object
const LastGenIndex = ""

// Indexes represents collection of indexes for specific object type
type Indexes struct {
	List map[string]*Index
}

// NameForStorable returns index value name for specific index and object
func (indexes *Indexes) NameForStorable(indexName string, storable runtime.Storable, codec Codec) string {
	if index, exist := indexes.List[indexName]; exist {
		return index.NameForStorable(storable, codec)
	} else {
		panic(fmt.Sprintf("trying to access non-existing indexName for kind %s: %s", storable.GetKind(), indexName))
	}
}

// NameForValue returns index value name for specific index, key and value
func (indexes *Indexes) NameForValue(indexName string, key runtime.Key, value interface{}, codec Codec) string {
	if index, exist := indexes.List[indexName]; exist {
		return index.NameForValue(key, value, codec)
	} else {
		panic(fmt.Sprintf("trying to access non-existing indexName for key %s: %s", key, indexName))
	}
}

var noopValueTransform = func(val interface{}) interface{} {
	return val
}

// IndexesFor returns (cached) collection of indexes for specified object typed
func IndexesFor(info *runtime.TypeInfo) *Indexes {
	indexCacheMu.Lock()
	defer indexCacheMu.Unlock()

	indexes, exist := indexCache[info.Kind]
	if !exist {
		indexes = &Indexes{map[string]*Index{}}
		indexCache[info.Kind] = indexes

		if info.Versioned {
			indexes.List[LastGenIndex] = &Index{
				Type: IndexTypeLastGen,
			}
		}

		t := reflect.TypeOf(info.New())
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("store")

			if strings.Contains(tag, "index") {
				// todo validate that field is accessible
				transformer := info.IndexValueTransforms[f.Name]
				if transformer == nil {
					transformer = noopValueTransform
				}
				indexes.List[f.Name] = &Index{
					Type:           IndexTypeListGen,
					Field:          f.Name,
					ValueTransform: transformer,
					rFieldId:       i,
				}
			}
		}
	}

	return indexes
}

type IndexType int

const (
	IndexTypeUndef IndexType = iota
	IndexTypeLastGen
	IndexTypeListGen
)

func (indexType IndexType) String() string {
	indexTypes := [...]string{
		"lastgen",
		"listgen",
	}

	if indexType < 1 || indexType > 2 {
		panic(fmt.Sprintf("unknown index type: %d", indexType))
	}

	return indexTypes[indexType-1]
}

type Index struct {
	Type           IndexType
	Field          string
	ValueTransform runtime.ValueTransform
	rFieldId       int
}

func (index *Index) NameForStorable(storable runtime.Storable, codec Codec) string {
	key := runtime.KeyForStorable(storable)

	if index.Type == IndexTypeLastGen {
		return index.NameForValue(key, nil, codec)
	}

	t := reflect.ValueOf(storable)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f := t.Field(index.rFieldId)

	return index.NameForValue(key, f.Interface(), codec)
}

func (index *Index) NameForValue(key runtime.Key, value interface{}, codec Codec) string {
	key = index.Type.String() + "/" + key
	if index.Type == IndexTypeLastGen {
		return key
	}

	value = index.ValueTransform(value)
	if value == nil {
		return ""
	}

	key += "/" + index.Field + "="

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
