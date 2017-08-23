package engine

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
)

// return cluster based on the "cluster" label. can return nil if there is no "cluster" label
func getCluster(policy *PolicyNamespace, labels LabelSet) (*Cluster, error) {
	var cluster *Cluster
	if clusterName, ok := labels.Labels["cluster"]; ok {
		if cluster, ok = policy.Clusters[clusterName]; !ok {
			return nil, fmt.Errorf("Cluster '%s' is not defined in policy", clusterName)
		}
	}
	return cluster, nil
}
