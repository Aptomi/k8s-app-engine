package language

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ClusterObject = &Info{
	Kind:        "cluster",
	Constructor: func() Base { return &Cluster{} },
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

// GetLabelSet returns a set of cluster labels
func (cluster *Cluster) GetLabelSet() LabelSet {
	return NewLabelSet(cluster.Labels)
}
