package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	"github.com/mattn/go-zglob"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func newApplyCommand(cfg *config.Client) *cobra.Command {
	paths := make([]string, 0)
	var wait bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply policy files",
		Long:  "apply policy files long",

		Run: func(cmd *cobra.Command, args []string) {
			allObjects, err := readFiles(paths)
			if err != nil {
				panic(fmt.Sprintf("Error while reading policy files for applying: %s", err))
			}

			client := rest.New(cfg, http.NewClient(cfg))
			result, err := client.Policy().Apply(allObjects)
			if err != nil {
				panic(fmt.Sprintf("Error while applying policy: %s", err))
			}

			// todo(slukjanov): replace with -o yaml / json / etc handler
			fmt.Println(result)

			if !wait {
				return
			}

			fmt.Println("Waiting for the first revision with updated policy to be applied")

			// todo limit retries by time ~ 10 mins
			for {
				// todo configurable
				time.Sleep(1 * time.Second)

				rev, revErr := client.Revision().ShowByPolicy(result.PolicyGeneration)
				if revErr != nil {
					fmt.Println("Error while getting revision for applied policy")
					continue
				}

				// todo print progress

				if rev.Progress.Finished {
					fmt.Println(rev)
					break
				}
			}
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files, dirs with policy to apply")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until first revision with updated policy will be fully applied")

	return cmd
}

func readFiles(policyPaths []string) ([]runtime.Object, error) {
	policyReg := runtime.NewRegistry().Append(lang.PolicyObjects...)
	codec := yaml.NewCodec(policyReg)

	files, err := findPolicyFiles(policyPaths)
	if err != nil {
		return nil, fmt.Errorf("error while searching for policy files: %s", err)
	}

	allObjects := make([]runtime.Object, 0)
	for _, file := range files {
		data, readErr := ioutil.ReadFile(file)
		if readErr != nil {
			return nil, fmt.Errorf("can't read file %s error: %s", file, readErr)
		}

		objects, decodeErr := codec.DecodeOneOrMany(data)
		if decodeErr != nil {
			return nil, fmt.Errorf("can't unmarshal file %s error: %s", file, decodeErr)
		}

		for _, obj := range objects {
			if !lang.IsPolicyObject(obj) {
				return nil, fmt.Errorf("only policy objects could be applied but got: %s", obj.GetKind())
			}
		}

		allObjects = append(allObjects, objects...)
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

	// todo(slukjanov): log list of files from which we're applying policy
	//fmt.Println("Apply policy from following files:")
	//for idx, policyPath := range allFiles {
	//	fmt.Println(idx, "-", policyPath)
	//}

	return allFiles, nil
}
