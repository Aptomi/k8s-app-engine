package dependency

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/io"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/spf13/cobra"
)

func newStatusCommand(cfg *config.Client) *cobra.Command {
	paths := make([]string, 0)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "dependency status",
		Long:  "dependency status",

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

			status, err := rest.New(cfg, http.NewClient(cfg)).Dependency().Status(dependencies, api.DependencyQueryDeploymentStatusOnly)
			if err != nil {
				panic(fmt.Sprintf("error while requesting dependency status: %s", err))
			}

			fmt.Println(status)
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files/dirs with dependency files")

	return cmd
}
