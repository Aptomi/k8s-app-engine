package language

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

var ClusterObject = &ObjectInfo{
	Kind:        "cluster",
	Constructor: func() BaseObject { return &Cluster{} },
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
	Cache struct {
		// store local proxy address when connection established (must stay private, starting from lowercase)
		tillerHost string

		// store kube external address (must stay private, starting from lowercase)
		kubeExternalAddress string

		// store istio svc name (must stay private, starting from lowercase)
		istioSvc string
	}
}

func (cluster *Cluster) SetTillerHost(tillerHost string) {
	cluster.Cache.tillerHost = tillerHost
}

func (cluster *Cluster) GetTillerHost() string {
	return cluster.Cache.tillerHost
}

func (cluster *Cluster) SetKubeExternalAddress(kubeExternalAddress string) {
	cluster.Cache.kubeExternalAddress = kubeExternalAddress
}

func (cluster *Cluster) GetKubeExternalAddress() string {
	return cluster.Cache.kubeExternalAddress
}

func (cluster *Cluster) SetIstioSvc(istioSvc string) {
	cluster.Cache.istioSvc = istioSvc
}

func (cluster *Cluster) GetIstioSvc() string {
	return cluster.Cache.istioSvc
}

// GetLabelSet returns a set of cluster labels
func (cluster *Cluster) GetLabelSet() LabelSet {
	return NewLabelSet(cluster.Labels)
}
