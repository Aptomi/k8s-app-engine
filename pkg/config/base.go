package config

import "github.com/sirupsen/logrus"

// Base is the interface for all configs used in Aptomi (e.g. client config, server config)
type Base interface {
	IsDebug() bool
	GetLogLevel() logrus.Level
}
