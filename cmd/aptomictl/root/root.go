package root

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/endpoints"
	"github.com/Aptomi/aptomi/cmd/aptomictl/gen"
	"github.com/Aptomi/aptomi/cmd/aptomictl/policy"
	"github.com/Aptomi/aptomi/cmd/aptomictl/revision"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/config"
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
	// Config is the global instance of the client config
	Config = &config.Client{}

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

	common.AddStringFlag(Command, "auth.username", "username", "u", "", EnvPrefix+"_USERNAME", "Username")
	common.AddDurationFlag(Command, "http.timeout", "timeout", "", 15*time.Second, EnvPrefix+"_TIMEOUT", "HTTP Timeout")

	// Add sub commands
	Command.AddCommand(
		common.NewVersionCommand(&Config.Output),
		endpoints.NewCommand(Config),
		policy.NewCommand(Config),
		revision.NewCommand(Config),
		gen.NewCommand(Config),
	)
}

func preRun(command *cobra.Command, args []string) {
	err := common.ReadConfig(viper.GetViper(), Config, defaultConfigDir())
	if err != nil {
		panic(fmt.Sprintf("error while loading config: %s", err))
	}
}

func defaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("can't find home dir: %s", err))
	}

	return path.Join(home, ".aptomi")
}
