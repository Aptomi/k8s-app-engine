package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

func newDeleteCommand(cfg *config.Client) *cobra.Command {
	paths := make([]string, 0)
	var wait bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete policy files",
		Long:  "delete policy files long",

		Run: func(cmd *cobra.Command, args []string) {
			allObjects, err := readLangFromFiles(paths)
			if err != nil {
				panic(fmt.Sprintf("Error while reading policy files for deleting: %s", err))
			}

			client := rest.New(cfg, http.NewClient(cfg))
			result, err := client.Policy().Delete(allObjects)
			if err != nil {
				panic(fmt.Sprintf("Error while deleting policy: %s", err))
			}

			data, err := common.Format(cfg, false, result)
			if err != nil {
				panic(fmt.Sprintf("Error while formating policy update result: %s", err))
			}
			fmt.Println(string(data))

			if !wait {
				return
			}

			waitForApplyToFinish(client, result)
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files, dirs with policy to delete")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until first revision with updated policy will be fully deleted")

	return cmd
}
