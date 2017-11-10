package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/spf13/cobra"
)

func newShowCommand(cfg *config.Client) *cobra.Command {
	var gen uint64 // == runtime.Generation

	cmd := &cobra.Command{
		Use:   "show",
		Short: "policy show",
		Long:  "policy show long",

		Run: func(cmd *cobra.Command, args []string) {
			result, err := rest.New(cfg, http.NewClient(cfg)).Policy().Show(runtime.Generation(gen))
			if err != nil {
				panic(fmt.Sprintf("Error while showing policy: %s", err))
			}

			// todo(slukjanov): replace with -o yaml / json / etc handler
			fmt.Println(result)
		},
	}

	cmd.Flags().Uint64VarP(&gen, "generation", "g", 0, "Policy generation")

	return cmd
}
