package language

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

func LoadUnitTestsPolicy(storeDir string) *Policy {
	loader := NewFileLoader(storeDir)

	policy := NewPolicy()
	objects, err := loader.LoadObjects()
	if err != nil {
		panic(fmt.Sprintf("Error while loading test policy: %s", err))
	}

	for _, object := range objects {
		policy.AddObject(object)
	}

	return policy
}

func NewFileLoader(path string) *FileLoader {
	catalog := NewObjectCatalog(ServiceObject, ContextObject, ClusterObject, RuleObject, DependencyObject)

	return &FileLoader{yaml.NewCodec(catalog), path}
}

type FileLoader struct {
	codec codec.MarshalUnmarshaler
	path  string
}

func (store *FileLoader) LoadObjects() ([]Base, error) {
	files, _ := zglob.Glob(filepath.Join(store.path, "**", "*.yaml"))
	sort.Strings(files)

	result := make([]Base, 0)
	for _, f := range files {
		if !strings.Contains(f, "external") {
			objects, err := store.loadObjectsFromFile(f)
			if err != nil {
				return nil, fmt.Errorf("Error while loading objects from file %s: %s", f, err)
			}

			result = append(result, objects...)
		}
	}

	// This hack is needed to make sure that we'll get test data in the same way like after marshaling objects
	// and storing them in DB. Example: empty fields will be stored anyway, while we omitting them in test data.
	data, err := store.codec.MarshalMany(result)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling loaded objects: %s", err)
	}
	result, err = store.codec.UnmarshalOneOrMany(data)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling loaded objects: %s", err)
	}

	return result, nil
}

func (store *FileLoader) loadObjectsFromFile(path string) ([]Base, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error while reading file %store: %store", path, err)
	}
	objects, err := store.codec.UnmarshalOneOrMany(data)
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshaling file %store: %store", path, err)
	}

	return objects, nil
}
