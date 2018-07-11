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

const LastGenIndex = ""

type Indexes struct {
	List map[string]*Index
}

func (indexes *Indexes) KeyForStorable(indexName string, storable runtime.Storable, codec Codec) string {
	if index, exist := indexes.List[indexName]; exist {
		return index.KeyForStorable(storable, codec)
	} else {
		panic(fmt.Sprintf("trying to access non-existing indexName for kind %s: %s", storable.GetKind(), indexName))
	}
}

func (indexes *Indexes) KeyForValue(indexName string, key runtime.Key, value interface{}, codec Codec) string {
	if index, exist := indexes.List[indexName]; exist {
		return index.KeyForValue(key, value, codec)
	} else {
		panic(fmt.Sprintf("trying to access non-existing indexName for key %s: %s", key, indexName))
	}
}

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

			// todo better tag, probably `index:"gen,list"`
			if strings.Contains(tag, "gen_index") {
				// todo validate that field is accessible
				// todo cache reflection objects

				indexes.List[f.Name] = &Index{
					Type:     IndexTypeListGen,
					Field:    f.Name,
					rFieldId: i,
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
	Type     IndexType
	Field    string
	rFieldId int
}

func (index *Index) KeyForStorable(storable runtime.Storable, codec Codec) string {
	key := runtime.KeyForStorable(storable)

	if index.Type == IndexTypeLastGen {
		return index.KeyForValue(key, nil, codec)
	}

	t := reflect.ValueOf(storable)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f := t.Field(index.rFieldId)

	return index.KeyForValue(key, f.Interface(), codec)
}

func (index *Index) KeyForValue(key runtime.Key, value interface{}, codec Codec) string {
	key = index.Type.String() + "/" + key
	if index.Type == IndexTypeLastGen {
		return key
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

func (list *IndexValueList) First() []byte {
	if len(*list) == 0 {
		return nil
	}

	return (*list)[0]
}

func (list *IndexValueList) Last() []byte {
	if len(*list) == 0 {
		return nil
	}

	return (*list)[len(*list)-1]
}
