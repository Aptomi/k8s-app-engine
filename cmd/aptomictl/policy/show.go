package policy

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

func newShowCommand(cfg *config.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "policy show",
		Long:  "policy show long",

		Run: func(cmd *cobra.Command, args []string) {
			//err := client.Show(cfg)
			//if err != nil {
			//	panic(fmt.Sprintf("Error while showing policy: %s", err))
			//}
		},
	}
}
