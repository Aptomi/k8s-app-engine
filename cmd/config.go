package cmd

import (
	"github.com/Aptomi/aptomi/pkg/slinga/config"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/yaml"
	log "github.com/Sirupsen/logrus"
	vp "github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

func ReadConfig(viper *vp.Viper, cfg config.Base, defaultConfigDir string) {
	configFilePath := viper.GetString("config")

	if configFilePath != "" { // if config path provided, use it and don't look for default locations
		configAbsPath, err := filepath.Abs(configFilePath)
		if err != nil {
			log.Panicf("Error getting abs path for %s: %s", configFilePath, err)
		}

		processConfigAbsPath(viper, configAbsPath)
	} else { // if no config path available, search in default places
		// if there is no default config file - just skip config parsing
		if !isConfigExists(defaultConfigDir) {
			log.Infof("Can't find config file in default config dir: %s", defaultConfigDir)
			return
		}

		viper.AddConfigPath(defaultConfigDir)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Can't read config: %s", err)
	}

	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Panicf("Unable to unmarshal config: %s", err)
	}

	if cfg.IsDebug() {
		log.SetLevel(log.DebugLevel)
	}

	log.Debugf("Config:\n%s", yaml.SerializeObject(cfg))
}

func processConfigAbsPath(viper *vp.Viper, path string) {
	if stat, err := os.Stat(path); err == nil {
		if stat.IsDir() { // if dir provided, use only it
			viper.AddConfigPath(path)
		} else { // if specific file provided, use only it
			viper.SetConfigFile(path)
		}
	} else if os.IsNotExist(err) {
		log.Panicf("Specified config path %s doesn't exists: %s", path, err)
	} else {
		log.Panicf("Error while processing specified config path %s: %s", path, err)
	}
}

func isConfigExists(configDir string) bool {
	exists := false

	// check all supported config types
	for _, supportedType := range vp.SupportedExts {
		defaultConfigFile := path.Join(configDir, "config."+supportedType)

		if _, err := os.Stat(defaultConfigFile); err == nil {
			exists = true
			break
		}
	}

	return exists
}
