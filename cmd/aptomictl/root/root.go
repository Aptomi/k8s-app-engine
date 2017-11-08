package root

import (
	"github.com/Aptomi/aptomi/cmd/aptomictl/endpoints"
	"github.com/Aptomi/aptomi/cmd/aptomictl/policy"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
)

const (
	EnvPrefix = "APTOMICTL"
)

var (
	Config  = &config.Client{}
	Command = &cobra.Command{
		Use:   "aptomictl",
		Short: "aptomictl controls Aptomi",
		Long:  "aptomictl controls Aptomi",

		PersistentPreRun: preRun,

		Run: func(cmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {
	viper.SetEnvPrefix(EnvPrefix)

	common.AddDefaultFlags(Command, EnvPrefix)

	// Add sub commands
	Command.AddCommand(
		common.Version,
		endpoints.NewCommand(Config),
		policy.NewCommand(Config),
	)
}

func preRun(command *cobra.Command, args []string) {
	common.ReadConfig(viper.GetViper(), Config, defaultConfigDir())
}

func defaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Panicf("Can't find homedir: %s", err)
	}

	return path.Join(home, ".aptomi")
}
