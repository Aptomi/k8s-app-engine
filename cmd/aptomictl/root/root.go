package root

import (
	"github.com/Aptomi/aptomi/cmd/aptomictl/dependency"
	"github.com/Aptomi/aptomi/cmd/aptomictl/gen"
	"github.com/Aptomi/aptomi/cmd/aptomictl/login"
	"github.com/Aptomi/aptomi/cmd/aptomictl/policy"
	"github.com/Aptomi/aptomi/cmd/aptomictl/revision"
	"github.com/Aptomi/aptomi/cmd/aptomictl/state"
	"github.com/Aptomi/aptomi/cmd/aptomictl/version"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
	"time"
)

const (
	// EnvPrefix is the prefix for all environment variables used by aptomictl
	EnvPrefix = "APTOMICTL"
)

var (
	// Config is the global instance of client config
	Config = &config.Client{}

	// ConfigFile is the path to config file used to read config
	ConfigFile = new(string)

	// Command is the main (root) cobra command for aptomictl
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

	common.AddStringFlag(Command, "output", "output", "o", "text", EnvPrefix+"_OUTPUT", "Output format. One of: text (default), json, yaml")

	common.AddDurationFlag(Command, "http.timeout", "timeout", "", 60*time.Second, EnvPrefix+"_TIMEOUT", "Specifies time limit for receiving a reply from the server")

	// Add sub commands
	Command.AddCommand(
		login.NewCommand(Config, ConfigFile),
		dependency.NewCommand(Config),
		policy.NewCommand(Config),
		revision.NewCommand(Config),
		state.NewCommand(Config),
		gen.NewCommand(Config),
		version.NewCommand(Config),
	)
}

func preRun(command *cobra.Command, args []string) {
	if command.Parent() != nil {
		err := common.ReadConfig(viper.GetViper(), Config, defaultConfigDir())
		if err != nil {
			log.Fatalf("error while loading config: %s", err)
		}

		usedConfigFile := viper.ConfigFileUsed()
		*ConfigFile = usedConfigFile

		log.Infof("Using config file: %s", usedConfigFile)
	}
}

func defaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("can't find home dir: %s", err)
	}

	return path.Join(home, ".aptomi")
}
