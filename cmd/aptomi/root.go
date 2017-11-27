package main

import (
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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
	common.AddStringFlag(aptomiCmd, "db.connection", "db", "", "/etc/aptomi/db.bolt", envPrefix+"_DB_CONN", "DB connection string")
	common.AddStringFlag(aptomiCmd, "ui.schema", "ui-schema", "", "http", envPrefix+"_SCHEMA", "Server UI schema")
	common.AddStringFlag(aptomiCmd, "ui.host", "ui-host", "", "127.0.0.1", envPrefix+"_HOST", "Server UI host")
	common.AddStringFlag(aptomiCmd, "ui.port", "ui-port", "", "8080", envPrefix+"_PORT", "Server UI port")

	aptomiCmd.AddCommand(common.Version)
}

func preRun(command *cobra.Command, args []string) {
	err := common.ReadConfig(viper.GetViper(), cfg, "/etc/aptomi")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
