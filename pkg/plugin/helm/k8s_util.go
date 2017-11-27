package helm

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	rbacapi "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

func (cache *clusterCache) newKubeClient() (kubernetes.Interface, error) {
	client, err := kubernetes.NewForConfig(cache.kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}

	return client, nil
}

func (cache *clusterCache) getKubeExternalAddress() (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if len(cache.externalAddress) > 0 {
		return cache.externalAddress, nil
	}

	client, err := cache.newKubeClient()
	if err != nil {
		return "", fmt.Errorf("error while creating k8s client to cluster %s: %s", cache.cluster.Name, err)
	}

	nodes, err := client.CoreV1().Nodes().List(meta.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found for k8s cluster %s, it's critical eror", cache.cluster.Name)
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
		return "", fmt.Errorf("couldn't find external IP for cluster: %s", cache.cluster.Name)
	}

	cache.externalAddress = addr

	return addr, nil
}

func (cache *clusterCache) ensureKubeNamespace(client kubernetes.Interface, namespace string) error {
	_, err := client.CoreV1().Namespaces().Get(namespace, meta.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		ns := &api.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: namespace,
			},
		}
		_, createErr := client.CoreV1().Namespaces().Create(ns)
		if createErr != nil {
			return createErr
		}
	}

	return nil
}

func (cache *clusterCache) createKubeServiceAccount(client kubernetes.Interface, namespace string) error {
	sa := &api.ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			Name: "tiller-" + namespace,
		},
	}
	_, err := client.CoreV1().ServiceAccounts(namespace).Create(sa)

	return err
}

func (cache *clusterCache) createKubeClusterRoleBinding(client kubernetes.Interface, namespace string) error {
	crb := &rbacapi.ClusterRoleBinding{
		ObjectMeta: meta.ObjectMeta{
			Name: "tiller-" + namespace,
		},
		RoleRef: rbacapi.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
		Subjects: []rbacapi.Subject{{
			Kind:      "ServiceAccount",
			Name:      "tiller-" + namespace,
			Namespace: namespace,
		}},
	}
	_, err := client.RbacV1beta1().ClusterRoleBindings().Create(crb)

	return err
}
