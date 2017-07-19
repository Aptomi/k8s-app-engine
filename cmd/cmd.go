package cmd

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	debug "github.com/Aptomi/aptomi/pkg/slinga/log"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "aptomi",
	Short: "Aptomi - policy & governance for microservices",
	Long:  `Aptomi - policy & governance for microservices`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		debug.SetDebugLevel(log.DebugLevel)
		debug.SetLogFileName(GetAptomiDebugLogName())
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
