package file

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

type FileStore struct {
	store.BaseStore

	path string
}

func (store *FileStore) Open(connection string) error {
	store.path = connection

	return nil
}

func (store *FileStore) loadObjects() (map[object.Key]object.BaseObject, error) {
	files, _ := zglob.Glob(filepath.Join(store.path, "**", "*.yaml"))
	sort.Strings(files)

	result := make(map[object.Key]object.BaseObject, 0)
	for _, f := range files {
		if !strings.Contains(f, "external") {
			objects, err := store.loadObjectsFromFile(f)
			if err != nil {
				return nil, fmt.Errorf("Error while loading objects from file %store: %store", f, err)
			}

			for _, obj := range objects {
				result[obj.GetKey()] = obj
			}
		}
	}

	return result, nil
}

func (store *FileStore) loadObjectsFromFile(path string) ([]object.BaseObject, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error while reading file %store: %store", path, err)
	}
	objects, err := store.Codec.UnmarshalMany(data)
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshaling file %store: %store", path, err)
	}

	return objects, nil
}
