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

func LoadUnitTestsPolicy(storeDir string) *PolicyNamespace {
	loader := NewFileLoader(storeDir)

	policy := NewPolicyNamespace()
	objects, err := loader.LoadObjects()
	if err != nil {
		panic("Error while loading test policy: " + err.Error())
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

func (store *FileLoader) LoadObjects() ([]BaseObject, error) {
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

func (store *FileLoader) loadObjectsFromFile(path string) ([]BaseObject, error) {
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
