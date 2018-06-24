package claim

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
		Short: "claim endpoints",
		Long:  "claim endpoints",

		Run: func(cmd *cobra.Command, args []string) {
			allObjects, err := io.ReadLangObjects(paths)
			if err != nil {
				panic(fmt.Sprintf("error while reading policy files: %s", err))
			}

			claims := []*lang.Claim{}
			for _, obj := range allObjects {
				if d, ok := obj.(*lang.Claim); ok {
					claims = append(claims, d)
				}
			}

			printEndpoints(cfg, claims)
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files/dirs with claim files")
	return cmd
}

// TODO: ideally we should use common.Format() here to support writing into json and yaml, but it doesn't blend too well with maps and sorted keys
func printEndpoints(cfg *config.Client, claims []*lang.Claim) {
	result, errAPI := rest.New(cfg, http.NewClient(cfg)).Claim().Status(claims, api.ClaimQueryDeploymentStatusOnly)
	if errAPI != nil {
		panic(fmt.Sprintf("error while requesting claim status: %s", errAPI))
	}

	table := uitable.New()
	table.MaxColWidth = 120
	table.Wrap = true
	table.AddRow("CLAIM", "COMPONENT", "ENDPOINT TYPE", "ENDPOINT URL")
	for _, dKey := range util.GetSortedStringKeys(result.Status) {
		first := true
		for _, cKey := range util.GetSortedStringKeys(result.Status[dKey].Endpoints) {
			for _, eType := range util.GetSortedStringKeys(result.Status[dKey].Endpoints[cKey]) {
				claimKeyStr := ""
				if first {
					claimKeyStr = dKey
					first = false
				}
				table.AddRow(claimKeyStr, cKey, eType, result.Status[dKey].Endpoints[cKey][eType])
			}
		}
	}
	fmt.Println(table)
}
