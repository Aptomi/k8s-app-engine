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

var (
	configPath string // path to the configPath dir or specific configPath file

	AptomiCtlCmd = &cobra.Command{
		Use:   "aptomictl", // todo(slukjanov)
		Short: "",          // todo(slukjanov)
		Long:  "",          // todo(slukjanov)

		// parse the configPath if one is provided, or use the defaults
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if configPath != "" { // if config path provided, use it and don't look for default locations
				configPath, err := filepath.Abs(configPath)
				if err != nil {
					panic(fmt.Sprintf("Error reading filepath: %s", err))
				}

				if stat, err := os.Stat(configPath); err == nil {
					if stat.IsDir() { // if dir provided, use only it
						viper.AddConfigPath(configPath)
					} else { // if specific file provided, use only it
						viper.SetConfigFile(configPath)
					}
				} else if os.IsNotExist(err) {
					panic(fmt.Sprintf("Path doesn't exists: %s error: %s", configPath, err))
				} else {
					panic(fmt.Sprintf("Error while processing path: %s", err))
				}

				if err := viper.ReadInConfig(); err != nil {
					panic(fmt.Sprintf("Failed to read config file: %s with error: %s", configPath, err))
				}
			} else { // if no config path provided, search in default places
				home, err := homedir.Dir()
				if err != nil {
					panic(fmt.Sprintf("Can't find homedir: %s", err))
				}

				// search config in home directory with name ".aptomi/config" (without extension).
				viper.AddConfigPath(path.Join(home, ".aptomi"))
				viper.SetConfigName("config")
			}

			// todo(slukjanov): if no config file found, it's okay, continue with defaults and CLI args

			if err := viper.ReadInConfig(); err != nil {
				panic(fmt.Sprintf("Can't read configPath: %s", err))
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {
	AptomiCtlCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file or dir path")
}
