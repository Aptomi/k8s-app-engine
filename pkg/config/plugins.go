package config

import "time"

// Plugins represents configs for all plugins
type Plugins struct {
	K8s    K8s
	K8sRaw K8sRaw
	Helm   Helm
}

// K8s represents config for Kubernetes cluster plugin
type K8s struct {
	Timeout time.Duration
}

// K8sRaw represents config for Kubernetes Raw code plugin
type K8sRaw struct {
	DataNamespace string
}

// Helm represents configs for Helm code plugin
type Helm struct {
	Timeout time.Duration
}
