package language

import "github.com/Frostman/aptomi/pkg/slinga/language/yaml"

// Cluster defines individual K8s cluster and way to access it
type Cluster struct {
	Name     string
	Type     string
	Labels   map[string]string
	Metadata struct {
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

// Loads cluster from file
func loadClusterFromFile(fileName string) *Cluster {
	return yaml.LoadObjectFromFile(fileName, new(Cluster)).(*Cluster)
}
