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

		// store local proxy address when connection established
		TillerHost string

		// store kube external address
		KubeExternalAddress string

		// store istio svc name
		IstioSvc string
	}
}

// Loads cluster from file
func loadClusterFromFile(fileName string) *Cluster {
	return yaml.LoadObjectFromFile(fileName, new(Cluster)).(*Cluster)
}
