package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	"github.com/spf13/cobra"
	"time"
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

			fmt.Println("Waiting for the first revision with updated policy to be deleted")

			var rev *engine.Revision
			interval := 5 * time.Second
			finished := retry.Do(60, interval, func() bool {
				var revErr error
				rev, revErr = client.Revision().ShowByPolicy(result.PolicyGeneration)
				if revErr != nil {
					fmt.Printf("Can't get revision for applied policy: %s, retrying in %s\n", revErr, interval)
					return false
				}

				fmt.Printf("Applying changes: %d out of %d total\n", rev.Progress.Current, rev.Progress.Total)
				return rev.Status != engine.RevisionStatusInProgress
			})

			if !finished {
				// todo pretty print
				fmt.Println("Wait for revision delete timedout", rev)
			} else if rev.Status == engine.RevisionStatusSuccess {
				// todo pretty print
				fmt.Println("Success! Policy deleted", rev)
			} else if rev.Status == engine.RevisionStatusError {
				// todo pretty print
				fmt.Println("Revision delete failed for policy", rev)
			}
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files, dirs with policy to delete")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until first revision with updated policy will be fully deleted")

	return cmd
}
