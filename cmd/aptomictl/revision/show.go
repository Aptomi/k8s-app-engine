package revision

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newShowCommand(cfg *config.Client) *cobra.Command {
	var gen uint64

	cmd := &cobra.Command{
		Use:   "show",
		Short: "revision show",
		Long:  "revision show long",

		Run: func(cmd *cobra.Command, args []string) {
			result, err := rest.New(cfg, http.NewClient(cfg)).Revision().Show(runtime.Generation(gen))

			if err != nil {
				log.Fatalf("error while showing revision: %s", err)
			}

			// todo(slukjanov): replace with -o yaml / json / etc handler
			fmt.Println(result)
		},
	}

	cmd.Flags().Uint64VarP(&gen, "generation", "g", 0, "Revision generation")

	return cmd
}
