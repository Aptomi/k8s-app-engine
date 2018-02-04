package state

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

func newResetCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "state reset",
		Long:  "state reset long",

		Run: func(cmd *cobra.Command, args []string) {
			rev, err := rest.New(cfg, http.NewClient(cfg)).State().Reset()

			if err != nil {
				panic(fmt.Sprintf("Error while showing revision: %s", err))
			}

			// todo(slukjanov): replace with -o yaml / json / etc handler
			fmt.Println(rev)
		},
	}

	return cmd
}
