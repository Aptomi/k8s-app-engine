package file

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	. "github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

type FileStore struct {
	BaseStore

	path string
}

func (store *FileStore) Open(connection string) error {
	store.path = connection

	return nil
}

func (store *FileStore) Close() error {
	// noop
	return nil
}

//func (store *FileStore) GetNewestOne(namespace string, kind Kind, name string) (BaseObject, error) {
//	objects, err := store.LoadObjects()
//	if err != nil {
//		return nil, fmt.Errorf("Can't load object with error: %s", err)
//	}
//
//	// todo impl normal search
//	if namespace != "system" && kind != "policy" && name != "main" {
//		panic("File store is intermediate solution, will be removed soon")
//	}
//
//	for _, object := range objects {
//		if object.GetKind() == kind {
//			return object, nil
//		}
//	}
//
//	return nil, fmt.Errorf("Can't find object namesapce=%s kind=%s name=%s", namespace, kind, name)
//}
//
//func (store *FileStore) GetManyByKeys(keys []Key) ([]BaseObject, error) {
//	result := make([]BaseObject, 0, len(keys))
//
//	objects, err := store.LoadObjects()
//	if err != nil {
//		return nil, fmt.Errorf("Can't load object with error: %s", err)
//	}
//
//	for _, key := range keys {
//		object, ok := objects[key]
//		if !ok {
//			return nil, fmt.Errorf("Object with key %s not found", key)
//		}
//		result = append(result, object)
//	}
//
//	return result, nil
//}

func (store *FileStore) LoadObjects() ([]BaseObject, error) {
	files, _ := zglob.Glob(filepath.Join(store.path, "**", "*.yaml"))
	sort.Strings(files)

	result := make([]BaseObject, 0)
	for _, f := range files {
		if !strings.Contains(f, "external") {
			objects, err := store.loadObjectsFromFile(f)
			if err != nil {
				return nil, fmt.Errorf("Error while loading objects from file %store: %store", f, err)
			}

			result = append(result, objects...)
		}
	}

	return result, nil
}

func (store *FileStore) loadObjectsFromFile(path string) ([]BaseObject, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error while reading file %store: %store", path, err)
	}
	objects, err := store.Codec.UnmarshalOneOrMany(data)
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshaling file %store: %store", path, err)
	}

	return objects, nil
}
