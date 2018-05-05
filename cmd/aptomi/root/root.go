package root

import (
	"github.com/Aptomi/aptomi/cmd/aptomi/server"
	"github.com/Aptomi/aptomi/cmd/aptomi/version"
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
	// Config is the global instance of server config
	Config = &config.Server{}

	// Command is the main (root) cobra command for aptomi
	Command = &cobra.Command{
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
	common.AddDefaultFlags(Command, envPrefix)

	// add server-specific flags
	common.AddStringFlag(Command, "db.connection", "db", "", "/var/lib/aptomi/db.bolt", envPrefix+"_DB_CONN", "DB connection string")
	common.AddStringFlag(Command, "ui.schema", "ui-schema", "", "http", envPrefix+"_SCHEMA", "Server UI schema")
	common.AddBoolFlag(Command, "ui.enable", "ui", "", true, envPrefix+"_UI", "Enable server to serve UI")
	common.AddDurationFlag(Command, "enforcer.interval", "enforcer-interval", "", 60*time.Second, envPrefix+"_ENFORCER_INTERVAL", "Enforcer interval")

	Command.AddCommand(
		version.NewVersionCommand(),
		server.NewServerCommand(Config),
	)
}

func preRun(command *cobra.Command, args []string) {
	if command.Parent() != nil {
		err := common.ReadConfig(viper.GetViper(), Config, "/etc/aptomi")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
}
