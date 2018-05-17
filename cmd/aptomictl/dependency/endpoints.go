package dependency

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/io"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

func newEndpointsCommand(cfg *config.Client) *cobra.Command {
	paths := make([]string, 0)

	cmd := &cobra.Command{
		Use:   "endpoints",
		Short: "dependency endpoints",
		Long:  "dependency endpoints",

		Run: func(cmd *cobra.Command, args []string) {
			allObjects, err := io.ReadLangObjects(paths)
			if err != nil {
				panic(fmt.Sprintf("error while reading policy files: %s", err))
			}

			dependencies := []*lang.Dependency{}
			for _, obj := range allObjects {
				if d, ok := obj.(*lang.Dependency); ok {
					dependencies = append(dependencies, d)
				}
			}

			printEndpoints(cfg, dependencies)
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files/dirs with dependency files")
	return cmd
}

// TODO: ideally we should use common.Format() here to support writing into json and yaml, but it doesn't blend too well with maps and sorted keys
func printEndpoints(cfg *config.Client, dependencies []*lang.Dependency) {
	result, errAPI := rest.New(cfg, http.NewClient(cfg)).Dependency().Status(dependencies, api.DependencyQueryDeploymentStatusOnly)
	if errAPI != nil {
		panic(fmt.Sprintf("error while requesting dependency status: %s", errAPI))
	}

	table := uitable.New()
	table.MaxColWidth = 120
	table.Wrap = true
	table.AddRow("DEPENDENCY", "COMPONENT", "ENDPOINT TYPE", "ENDPOINT URL")
	for _, dKey := range util.GetSortedStringKeys(result.Status) {
		first := true
		for _, cKey := range util.GetSortedStringKeys(result.Status[dKey].Endpoints) {
			for _, eType := range util.GetSortedStringKeys(result.Status[dKey].Endpoints[cKey]) {
				depKeyStr := ""
				if first {
					depKeyStr = dKey
					first = false
				}
				table.AddRow(depKeyStr, cKey, eType, result.Status[dKey].Endpoints[cKey][eType])
			}
		}
	}
	fmt.Println(table)
}
