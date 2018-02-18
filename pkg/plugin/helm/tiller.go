package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	rbacapi "k8s.io/client-go/pkg/apis/rbac/v1beta1"
	"k8s.io/helm/cmd/helm/installer"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"strings"
	"time"
)

func (plugin *Plugin) ensureTillerTunnel(eventLog *event.Log) error {
	client, clientErr := plugin.kube.NewClient()
	if clientErr != nil {
		return clientErr
	}

	// we should be able to list pods in tiller namespace
	_, clientErr = client.CoreV1().Pods(plugin.tillerNamespace).List(meta.ListOptions{})
	if clientErr != nil {
		return fmt.Errorf("error while pre-flight check for cluster %s: %s", plugin.cluster.Name, clientErr)
	}

	eventLog.WithFields(event.Fields{}).Debugf("Creating k8s tunnel for cluster %s", plugin.cluster.Name)

	var tunnelErr error
	ok := retry.Do(120, 5*time.Second, func() bool {
		if plugin.tillerTunnel != nil {
			plugin.tillerTunnel.Close()
		}
		plugin.tillerTunnel, tunnelErr = portforwarder.New(plugin.tillerNamespace, client, plugin.kube.KubeConfig)

		if tunnelErr != nil {
			if strings.Contains(tunnelErr.Error(), "could not find tiller") {
				tillerErr := plugin.setupTiller(client, eventLog)
				if tillerErr != nil {
					tunnelErr = tillerErr
				} else {
					// if no error, let's try open tunnel again
					return false
				}
			}

			eventLog.WithFields(event.Fields{}).Debugf("Retrying after error while creating k8s tunnel for cluster %s: %s", plugin.cluster.Name, tunnelErr)

			return false
		}

		port := plugin.tillerTunnel.Local
		plugin.tillerHost = fmt.Sprintf("localhost:%d", port)

		helmClient, err := plugin.newClient()
		if err != nil {
			tunnelErr = fmt.Errorf("can't create helm client for just created k8s tunnel for cluster %s: %s", plugin.cluster.Name, err)
			eventLog.WithFields(event.Fields{}).Debugf("Retrying after error: %s", tunnelErr)
			return false
		}

		_, err = helmClient.ListReleases(helm.ReleaseListLimit(1))
		if err != nil {
			tunnelErr = fmt.Errorf("can't do helm list using just created k8s tunnel for cluster %s: %s", plugin.cluster.Name, err)
			eventLog.WithFields(event.Fields{}).Debugf("Retrying after error: %s", tunnelErr)
			return false
		}

		eventLog.WithFields(event.Fields{}).Debugf("Created k8s tunnel using local port %d for cluster %s", port, plugin.cluster.Name)

		return true
	})

	if !ok {
		if tunnelErr != nil {
			return tunnelErr
		}

		return fmt.Errorf("tiller tunnel creation timeout for cluster: %s", plugin.cluster.Name)
	}

	return nil
}

func (plugin *Plugin) setupTiller(client kubernetes.Interface, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Debugf("Setting up tiller in cluster %s namespace %s", plugin.cluster.Name, plugin.tillerNamespace)

	err := plugin.kube.EnsureNamespace(client, plugin.tillerNamespace)
	if err != nil {
		return err
	}

	saName := "tiller-" + plugin.tillerNamespace
	err = ensureKubeServiceAccount(client, plugin.tillerNamespace, saName)
	if err != nil {
		return err
	}

	err = ensureKubeAdminClusterRoleBinding(client, plugin.tillerNamespace, saName)
	if err != nil {
		return err
	}

	return installer.Install(client, &installer.Options{
		Namespace:      plugin.tillerNamespace,
		ImageSpec:      "gcr.io/kubernetes-helm/tiller:v2.6.2",
		ServiceAccount: saName,
	})
}

func ensureKubeServiceAccount(client kubernetes.Interface, namespace string, name string) error {
	_, err := client.CoreV1().ServiceAccounts(namespace).Get(name, meta.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		sa := &api.ServiceAccount{
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		}
		_, createErr := client.CoreV1().ServiceAccounts(namespace).Create(sa)
		return createErr
	}

	return err
}

func ensureKubeAdminClusterRoleBinding(client kubernetes.Interface, namespace string, name string) error {
	_, err := client.RbacV1beta1().ClusterRoleBindings().Get(name, meta.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		crb := &rbacapi.ClusterRoleBinding{
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
			RoleRef: rbacapi.RoleRef{
				Kind: "ClusterRole",
				Name: "cluster-admin",
			},
			Subjects: []rbacapi.Subject{{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			}},
		}
		_, createErr := client.RbacV1beta1().ClusterRoleBindings().Create(crb)
		return createErr
	}

	return err
}
