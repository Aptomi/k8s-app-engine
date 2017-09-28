package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ClusterObject = &object.Info{
	Kind:        "cluster",
	Versioned:   true,
	Constructor: func() object.Base { return &Cluster{} },
}

// Cluster defines individual K8s cluster and way to access it
type Cluster struct {
	Metadata

	Type   string
	Labels map[string]string
	Config struct {
		KubeContext     string
		TillerNamespace string
		Namespace       string
	}
}
