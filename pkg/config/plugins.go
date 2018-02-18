package config

import "time"

// Plugins represents configs for all plugins
type Plugins struct {
	Kube Kube
	Helm Helm
}

// Kube represents config for Kubernetes plugin
type Kube struct {
	Timeout time.Duration
}

// Helm represents configs for Helm plugin
type Helm struct {
	Timeout time.Duration
}
