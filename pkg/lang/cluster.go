package lang

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"gopkg.in/yaml.v2"
)

// TypeCluster is an informational data structure with Kind and Constructor for Cluster
var TypeCluster = &runtime.TypeInfo{
	Kind:        "cluster",
	Storable:    true,
	Versioned:   true,
	Constructor: func() runtime.Object { return &Cluster{} },
}

// Cluster defines an individual cluster where containers get deployed.
// Various cloud providers are supported via setting a cluster type (k8s, Amazon ECS, GKE, etc).
type Cluster struct {
	runtime.TypeKind `yaml:",inline"`
	Metadata         `validate:"required"`

	// Type is a cluster type. Based on its type, the appropriate deployment plugin will be called to deploy containers.
	Type string `validate:"clustertype"`

	// Labels is a set of labels attached to the cluster
	Labels map[string]string `yaml:"labels,omitempty" validate:"omitempty,labels"`

	// Config for a given cluster type
	Config interface{} `validate:"required"`
}

// ParseConfigInto parses cluster config into provided object
func (cluster *Cluster) ParseConfigInto(obj interface{}) error {
	data, err := yaml.Marshal(cluster.Config)
	if err != nil {
		return fmt.Errorf("error while marshalling cluster config into bytes using yaml: %s", err)
	}

	err = yaml.Unmarshal(data, obj)
	if err != nil {
		return fmt.Errorf("error while unmarshalling cluster config into provided object: %s", err)
	}

	return nil
}

// MakeCopy makes a shallow copy of the Cluster struct
func (cluster *Cluster) MakeCopy() *Cluster {
	return &Cluster{
		TypeKind: cluster.TypeKind,
		Metadata: cluster.Metadata,
		Type:     cluster.Type,
		Labels:   cluster.Labels,
		Config:   cluster.Config,
	}
}
