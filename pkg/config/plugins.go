package config

import "time"

// Plugins represents configs for all plugins
type Plugins struct {
	Helm Helm
}

// Helm represents configs for Helm plugin
type Helm struct {
	Timeout time.Duration
}
