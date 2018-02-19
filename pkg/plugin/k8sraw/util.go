package k8sraw

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"strings"
)

func (plugin *Plugin) storeManifest(client kubernetes.Interface, deployName, manifest string) error {
	namespace := "aptomi"
	name := strings.ToLower(util.EscapeName(fmt.Sprintf("aptomi-raw-%s", deployName)))

	cm, err := client.CoreV1().ConfigMaps(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			cm = &api.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name: name,
				},
				Data: map[string]string{
					"manifest": manifest,
				},
			}

			_, err = client.CoreV1().ConfigMaps(namespace).Create(cm)
		}

		return err
	}

	cm.Data = map[string]string{
		"manifest": manifest,
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Update(cm)

	return err
}

func (plugin *Plugin) loadManifest(client kubernetes.Interface, deployName string) (string, error) {
	namespace := "aptomi"
	name := strings.ToLower(util.EscapeName(fmt.Sprintf("aptomi-raw-%s", deployName)))

	cm, err := client.CoreV1().ConfigMaps(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("can't find data for deployment %s (should be stored in namespace %s): %s", deployName, namespace, err)
		}
		return "", err
	}

	manifest := cm.Data["manifest"]
	if len(manifest) == 0 {
		return "", fmt.Errorf("no manifest found in data for deployment %s (stored in configmap %s/%s", deployName, namespace, name)
	}

	return manifest, nil
}

func (plugin *Plugin) deleteManifest(client kubernetes.Interface, deployName string) error {
	namespace := "aptomi"
	name := strings.ToLower(util.EscapeName(fmt.Sprintf("aptomi-raw-%s", deployName)))

	err := client.CoreV1().ConfigMaps(namespace).Delete(name, &meta.DeleteOptions{})
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	return err
}
