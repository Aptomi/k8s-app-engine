package lang

// APIPolicy is a Policy representation for API filtered for specific user
type APIPolicy struct {
	Namespace map[string]*APIPolicyNamespace
}

// APIPolicyNamespace is a PolicyNamespace representation for API filtered for specific user
type APIPolicyNamespace struct {
	Services     map[string]*Service
	Contracts    map[string]*Contract
	Clusters     map[string]*Cluster
	Rules        map[string]*Rule
	ACLRules     map[string]*Rule
	Dependencies map[string]*Dependency
}

// APIPolicy returns Policy representation for API filtered for specific user
func (policy *Policy) APIPolicy() *APIPolicy {
	// TODO; implement
	return &APIPolicy{}
}

// APIPolicy returns Policy representation for API filtered for specific user
func (view *PolicyView) APIPolicy() *APIPolicy {
	result := view.Policy.APIPolicy()
	for k := range result.Namespace {
		result.Namespace[k].Clusters = view.filterClusters(result.Namespace[k].Clusters)
	}
	return result
}

// filterClusters returns clusters filtered for specific user
func (view *PolicyView) filterClusters(clusters map[string]*Cluster) map[string]*Cluster {
	result := make(map[string]*Cluster)
	for k := range clusters {
		filteredCluster := view.filterCluster(clusters[k])
		if filteredCluster != nil {
			result[k] = filteredCluster
		}
	}
	return result
}

// filterClusters returns user's view of the cluster (without any configuration parameters for non-admins)
func (view *PolicyView) filterCluster(cluster *Cluster) *Cluster {
	if view.ManageObject(cluster) == nil {
		// if user can manage cluster, return full information
		return cluster
	}

	if view.ViewObject(cluster) == nil {
		// if user can only view cluster, return stripped down information about the cluster
		result := cluster.MakeCopy()
		result.Config = ClusterConfig{}
		return result
	}

	// if user has no access, do not return anything
	return nil
}
