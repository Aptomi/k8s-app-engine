package main

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

const (
	envPrefix = "APTOMI"
)

var (
	aptomiCtlCmd = &cobra.Command{
		Use:   "aptomictl", // todo(slukjanov)
		Short: "",          // todo(slukjanov)
		Long:  "",          // todo(slukjanov)

		// parse the configPath if one is provided, or use the defaults
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			readConfig(viper.GetString("config"))

			// todo(slukjanov): pretty print all final configs
			fmt.Println(viper.AllSettings())
		},

		Run: func(cmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {
	viper.SetEnvPrefix(envPrefix)

	aptomiCtlCmd.PersistentFlags().StringP("config", "c", "", "Config file or dir path")

	aptomiCtlCmd.PersistentFlags().String("host", "127.0.0.1", "Server API host")
	err := viper.BindPFlag("server.host", aptomiCtlCmd.PersistentFlags().Lookup("host"))
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}
	err = viper.BindEnv("server.host", envPrefix+"_HOST")
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}

	aptomiCtlCmd.PersistentFlags().Uint16P("port", "p", 27866, "Server API port")
	err = viper.BindPFlag("server.port", aptomiCtlCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}
	err = viper.BindEnv("server.port", envPrefix+"_PORT")
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}
}

func readConfig(configFilePath string) {
	if configFilePath != "" { // if config path provided, use it and don't look for default locations
		configAbsPath, err := filepath.Abs(configFilePath)
		if err != nil {
			panic(fmt.Sprintf("Error getting abs path for %s error: %s", configFilePath, err))
		}

		if stat, err := os.Stat(configAbsPath); err == nil {
			if stat.IsDir() { // if dir provided, use only it
				viper.AddConfigPath(configAbsPath)
			} else { // if specific file provided, use only it
				viper.SetConfigFile(configAbsPath)
			}
		} else if os.IsNotExist(err) {
			panic(fmt.Sprintf("Path doesn't exists: %s error: %s", configAbsPath, err))
		} else {
			panic(fmt.Sprintf("Error while processing path: %s", err))
		}
	} else { // if no config path available, search in default places
		home, err := homedir.Dir()
		if err != nil {
			panic(fmt.Sprintf("Can't find homedir: %s", err))
		}

		defaultConfigDir := path.Join(home, ".aptomi")

		// check all supported config types
		defaultExists := false
		for _, supportedType := range viper.SupportedExts {
			defaultConfigFile := path.Join(defaultConfigDir, "config."+supportedType)

			// if there is no default config file - just skip config parsing
			if _, err := os.Stat(defaultConfigFile); err == nil {
				defaultExists = true
				break
			}
		}

		if !defaultExists {
			// todo(slukjanov): print some log message?
			return
		}

		// search config in home directory with name ".aptomi/config" (without extension).
		viper.AddConfigPath(defaultConfigDir)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Can't read config: %s", err))
	}
}
