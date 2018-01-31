package common

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand returns instance of cobra command that shows version from version package (injected at build tome)
func NewVersionCommand(cfg *config.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Aptomi Client version",
		Run: func(cmd *cobra.Command, args []string) {
			info := version.GetBuildInfo()

			data, err := Format(cfg, false, &info)
			if err != nil {
				panic(fmt.Sprintf("Error while formating policy: %s", err))
			}
			fmt.Println(string(data))

			fmt.Printf("Aptomi Client version: %s\n       git commit: %s\n       built: %s\n", info.GitVersion, info.GitCommit, info.BuildDate)
		},
	}
}
