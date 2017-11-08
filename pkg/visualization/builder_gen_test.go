package visualization

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/external/secrets"
	"github.com/Aptomi/aptomi/pkg/external/users"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	"github.com/mattn/go-zglob"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var integrationTestsLDAP = config.LDAP{
	Host:         "localhost",
	Port:         10389,
	BaseDN:       "o=aptomiOrg",
	Filter:       "(&(objectClass=organizationalPerson))",
	FilterByName: "(&(objectClass=organizationalPerson)(cn=%s))",
	LabelToAttributes: map[string]string{
		"name":              "cn",
		"description":       "description",
		"global_ops":        "isglobalops",
		"is_operator":       "isoperator",
		"mail":              "mail",
		"team":              "team",
		"org":               "o",
		"short-description": "role",
		"deactivated":       "deactivated",
	},
}

func TestVis(t *testing.T) {
	reg := runtime.NewRegistry().Append(api.Objects...)
	cod := yaml.NewCodec(reg)

	dir := "../../examples/03-twitter-analytics"
	allObjects, err := readFiles([]string{dir + "/policy"}, cod)
	extData := external.NewData(
		users.NewUserLoaderFromLDAP(integrationTestsLDAP, nil),
		secrets.NewSecretLoaderFromDir(dir+"/_external/secrets"),
	)
	if err != nil {
		panic(err)
	}

	policy := lang.NewPolicy()
	for _, obj := range allObjects {
		_ = policy.AddObject(obj)
	}

	resolver := resolve.NewPolicyResolver(policy, extData)
	resolution, eventLog, err := resolver.ResolveAllDependencies()
	if err != nil {
		eventLog.Save(&event.HookConsole{})
		panic(err)
	}

	// 1 - resolution graph
	graph := NewGraphBuilder(policy, resolution, extData).DependencyResolution(DependencyResolutionCfgDefault)
	graph.Save()

	// 2 - policy graph
	// graph := NewGraphBuilder(policy, resolution, extData).Policy(PolicyCfgDefault)
	// graph.Save()
}

func readFiles(policyPaths []string, codec runtime.Codec) ([]lang.Base, error) {
	files, err := findPolicyFiles(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("error while searching for policy files: %s", err)
	}

	allObjects := make([]lang.Base, 0)
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("can't read file %s error: %s", file, err)
		}

		objects, err := codec.DecodeOneOrMany(data)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal file %s error: %s", file, err)
		}

		for _, obj := range objects {
			langObj, ok := obj.(lang.Base)
			if !ok {
				return nil, fmt.Errorf("lang objects expected in file %s, but get: %s", file, obj.GetKind())
			}

			allObjects = append(allObjects, langObj)
		}
	}

	if len(allObjects) == 0 {
		return nil, fmt.Errorf("no objects found in %s", policyPaths)
	}

	return allObjects, nil
}

func findPolicyFiles(policyPaths []string) ([]string, error) {
	allFiles := make([]string, 0, len(policyPaths))

	for _, rawPolicyPath := range policyPaths {
		policyPath, errPath := filepath.Abs(rawPolicyPath)
		if errPath != nil {
			return nil, fmt.Errorf("error reading filepath: %s", errPath)
		}

		if stat, err := os.Stat(policyPath); err == nil {
			if stat.IsDir() { // if dir provided, use all yaml files from it
				files, errGlob := zglob.Glob(filepath.Join(policyPath, "**", "*.yaml"))
				if errGlob != nil {
					return nil, fmt.Errorf("error while searching yaml files in directory: %s error: %s", policyPath, err)
				}
				allFiles = append(allFiles, files...)
			} else { // if specific file provided, use it
				allFiles = append(allFiles, policyPath)
			}
		} else if os.IsNotExist(err) {
			return nil, fmt.Errorf("path doesn't exist: %s error: %s", policyPath, err)
		} else {
			return nil, fmt.Errorf("error while processing path: %s", err)
		}
	}

	sort.Strings(allFiles)

	//fmt.Println("Apply policy from following files:")
	//for idx, policyPath := range allFiles {
	//	fmt.Println(idx, "-", policyPath)
	//}

	return allFiles, nil
}
