package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"strings"
)

// Target represents a deployment target in Aptomi
type Target struct {
	// ClusterNamespace is a namespace in Aptomi, to which the cluster belongs
	ClusterNamespace string

	// ClusterName is a cluster name in Aptomi
	ClusterName string

	// Suffix is an additional specifier (e.g. k8s namespace in case of Helm and k8s plugins)
	Suffix string
}

// NewTarget creates a new deployment target, given a string in form [aptomi_namespace/]cluster[.suffix] (where suffix is typically a k8s namespace)
func NewTarget(target string) *Target {
	result := &Target{}

	// cut namespace
	{
		parts := strings.SplitN(target, "/", 2)
		if len(parts) >= 2 {
			result.ClusterNamespace = parts[0]
			target = parts[1]
		}
	}

	// split into cluster and suffix
	{
		parts := strings.SplitN(target, ".", 2)
		if len(parts) >= 2 {
			result.ClusterName = parts[0]
			result.Suffix = parts[1]
		} else {
			result.ClusterName = target
		}
	}

	return result
}

// GetCluster allows to look up a cluster, given a deployment target
func (target *Target) GetCluster(policy *Policy, currentNs string) (*Cluster, error) {
	var clusterObj runtime.Object
	var err error
	if len(target.ClusterNamespace) > 0 {
		// in specified namespace
		clusterObj, err = policy.GetObject(ClusterObject.Kind, target.ClusterName, target.ClusterNamespace)
		if err != nil {
			return nil, err
		}

		if clusterObj == nil {
			return nil, fmt.Errorf("cluster '%s/%s' not found", target.ClusterNamespace, target.ClusterName)
		}
	} else {
		// in current namespace
		clusterObj, err = policy.GetObject(ClusterObject.Kind, target.ClusterName, currentNs)
		if err != nil {
			return nil, err
		}

		// in system namespace
		if clusterObj == nil {
			clusterObj, err = policy.GetObject(ClusterObject.Kind, target.ClusterName, runtime.SystemNS)
			if err != nil {
				return nil, err
			}
		}

		if clusterObj == nil {
			return nil, fmt.Errorf("cluster '%s' not found (tried '%s' and '%s' namespaces)", target.ClusterName, currentNs, runtime.SystemNS)
		}
	}

	return clusterObj.(*Cluster), nil
}
