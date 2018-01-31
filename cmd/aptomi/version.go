package main

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand returns instance of cobra command that shows version from version package (injected at build tome)
func NewVersionCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the Aptomi Client version",
		Run: func(cmd *cobra.Command, args []string) {
			info := version.GetBuildInfo()

			data, err := common.Format(output, false, info)
			if err != nil {
				panic(fmt.Sprintf("Error while formating policy: %s", err))
			}
			fmt.Println(string(data))
		},
	}

	cmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format. One of: text (default), json, yaml")

	return cmd
}
