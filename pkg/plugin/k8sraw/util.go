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

func (plugin *Plugin) manifestConfigMapName(deployName string) string {
	return strings.ToLower(util.EscapeName(fmt.Sprintf("aptomi-raw-%s-%s", plugin.cluster.Name, deployName)))
}

func (plugin *Plugin) storeManifest(client kubernetes.Interface, deployName, manifest string) error {
	name := plugin.manifestConfigMapName(deployName)

	cm, err := client.CoreV1().ConfigMaps(plugin.dataNamespace).Get(name, meta.GetOptions{})
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

			_, err = client.CoreV1().ConfigMaps(plugin.dataNamespace).Create(cm)
		}

		return err
	}

	cm.Data = map[string]string{
		"manifest": manifest,
	}

	_, err = client.CoreV1().ConfigMaps(plugin.dataNamespace).Update(cm)

	return err
}

func (plugin *Plugin) loadManifest(client kubernetes.Interface, deployName string) (string, error) {
	name := plugin.manifestConfigMapName(deployName)

	cm, err := client.CoreV1().ConfigMaps(plugin.dataNamespace).Get(name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("can't find data for deployment %s (should be stored in namespace %s): %s", deployName, plugin.dataNamespace, err)
		}
		return "", err
	}

	manifest := cm.Data["manifest"]
	if len(manifest) == 0 {
		return "", fmt.Errorf("no manifest found in data for deployment %s (stored in configmap %s/%s", deployName, plugin.dataNamespace, name)
	}

	return manifest, nil
}

func (plugin *Plugin) deleteManifest(client kubernetes.Interface, deployName string) error {
	name := plugin.manifestConfigMapName(deployName)

	err := client.CoreV1().ConfigMaps(plugin.dataNamespace).Delete(name, &meta.DeleteOptions{})
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	return err
}
