package revision

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/spf13/cobra"
)

func newShowCommand(cfg *config.Client) *cobra.Command {
	var gen, policyGen uint64

	cmd := &cobra.Command{
		Use:   "show",
		Short: "revision show",
		Long:  "revision show long",

		Run: func(cmd *cobra.Command, args []string) {
			if gen != 0 && policyGen != 0 {
				panic(fmt.Sprintf("Only one of generation and policy generation could be used at the same time"))
			}

			var result *engine.Revision
			var err error

			if policyGen != 0 {
				result, err = rest.New(cfg, http.NewClient(cfg)).Revision().ShowByPolicy(runtime.Generation(policyGen))
			} else {
				result, err = rest.New(cfg, http.NewClient(cfg)).Revision().Show(runtime.Generation(gen))
			}

			if err != nil {
				panic(fmt.Sprintf("Error while showing revision: %s", err))
			}

			// todo(slukjanov): replace with -o yaml / json / etc handler
			fmt.Println(result)
		},
	}

	cmd.Flags().Uint64VarP(&gen, "generation", "g", 0, "Revision generation")
	cmd.Flags().Uint64VarP(&policyGen, "policy", "p", 0, "Policy generation")

	return cmd
}
