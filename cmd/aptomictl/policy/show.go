package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
	log "github.com/Sirupsen/logrus"
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
				log.Fatalf("error while showing policy: %s", err)
			}

			data, err := common.Format(cfg.Output, false, result)
			if err != nil {
				log.Fatalf("error while formatting policy: %s", err)
			}
			fmt.Println(string(data))
		},
	}

	cmd.Flags().Uint64VarP(&gen, "generation", "g", 0, "Policy generation")

	return cmd
}
