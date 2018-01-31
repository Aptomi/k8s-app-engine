package main

import (
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	envPrefix = "APTOMI"
)

var (
	cfg       = &config.Server{}
	aptomiCmd = &cobra.Command{
		Use:   "aptomi",
		Short: "Aptomi server",
		Long:  "Aptomi server",

		PersistentPreRun: preRun,

		Run: func(cmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {
	viper.SetEnvPrefix(envPrefix)

	// add common flags (shared between server and client)
	common.AddDefaultFlags(aptomiCmd, envPrefix)

	// add server-specific flags
	common.AddStringFlag(aptomiCmd, "db.connection", "db", "", "/var/lib/aptomi/db.bolt", envPrefix+"_DB_CONN", "DB connection string")
	common.AddStringFlag(aptomiCmd, "ui.schema", "ui-schema", "", "http", envPrefix+"_SCHEMA", "Server UI schema")
	common.AddBoolFlag(aptomiCmd, "ui.enable", "ui", "", true, envPrefix+"_UI", "Enable server to serve UI")
	common.AddDurationFlag(aptomiCmd, "enforcer.interval", "enforcer-interval", "", 5*time.Second, envPrefix+"_ENFORCER_INTERVAL", "Enforcer interval")

	aptomiCmd.AddCommand(NewVersionCommand())
}

func preRun(command *cobra.Command, args []string) {
	err := common.ReadConfig(viper.GetViper(), cfg, "/etc/aptomi")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
