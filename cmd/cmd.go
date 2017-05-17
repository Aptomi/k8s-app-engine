package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"flag"
)

var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "aptomi",
	Short: "Aptomi - policy & governance for microservices",
	Long:  `Aptomi - policy & governance for microservices`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flag.Parse()
		if verbose {
			flag.Lookup("logtostderr").Value.Set("true")
		}
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags for the application
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output (enables logging)")
}
