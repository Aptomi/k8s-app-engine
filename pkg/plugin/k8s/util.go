package k8s

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
)

// NewClient returns new instance of the Kubernetes client created from the cached in the plugin cluster config
func (plugin *Plugin) NewClient() (kubernetes.Interface, error) {
	client, err := kubernetes.NewForConfig(plugin.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating kubernetes client: %s", err)
	}

	return client, nil
}

// EnsureNamespace ensures configured Kubernetes namespace
func (plugin *Plugin) EnsureNamespace(client kubernetes.Interface, namespace string) error {
	_, err := client.CoreV1().Namespaces().Get(namespace, meta.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		ns := &api.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: namespace,
			},
		}
		_, createErr := client.CoreV1().Namespaces().Create(ns)
		return createErr
	}

	return err
}

func (plugin *Plugin) AddEndpointsFromService(service *api.Service, endpoints map[string]string) {
	// todo(slukjanov): support not only node ports
	if service.Spec.Type == "NodePort" {
		for _, port := range service.Spec.Ports {
			sURL := fmt.Sprintf("%s:%d", plugin.ExternalAddress, port.NodePort)

			// todo(slukjanov): could we somehow detect real schema? I think no :(
			if util.StringContainsAny(port.Name, "https") {
				sURL = "https://" + sURL
			} else if util.StringContainsAny(port.Name, "ui", "rest", "http", "grafana") {
				sURL = "http://" + sURL
			}

			name := port.Name
			if len(name) == 0 {
				name = port.TargetPort.String()
			}

			endpoints[name] = sURL
		}
	}
}

func (plugin *Plugin) getExternalAddress() (string, error) {
	client, err := plugin.NewClient()
	if err != nil {
		return "", fmt.Errorf("error while creating k8s client to cluster %s: %s", plugin.Cluster.Name, err)
	}

	nodes, err := client.CoreV1().Nodes().List(meta.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found for k8s cluster %s, it's critical eror", plugin.Cluster.Name)
	}

	returnFirst := func(addrType api.NodeAddressType) string {
		for _, node := range nodes.Items {
			for _, addr := range node.Status.Addresses {
				if addr.Type == addrType {
					return addr.Address
				}
			}
		}
		return ""
	}

	addr := returnFirst(api.NodeExternalIP)
	if addr == "" {
		addr = returnFirst(api.NodeInternalIP)
	}
	if addr == "" {
		addr = returnFirst(api.NodeHostName)
	}
	if addr == "" {
		return "", fmt.Errorf("couldn't find external IP for cluster: %s", plugin.Cluster.Name)
	}

	return addr, nil
}
