package k8sraw

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/api/core/v1"
	"strings"
)

var (
	configMapNameReplacer = strings.NewReplacer("#", "-", "_", "-")
)

func (p *Plugin) getManifestConfigMapName(deployName string) string {
	return strings.ToLower(configMapNameReplacer.Replace(fmt.Sprintf("aptomi-raw-%s-%s", p.cluster.Name, deployName)))
}

func (p *Plugin) storeManifest(client kubernetes.Interface, deployName, manifest string) error {
	name := p.getManifestConfigMapName(deployName)

	cm, err := client.CoreV1().ConfigMaps(p.dataNamespace).Get(name, meta.GetOptions{})
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

			_, err = client.CoreV1().ConfigMaps(p.dataNamespace).Create(cm)
		}

		return err
	}

	cm.Data = map[string]string{
		"manifest": manifest,
	}

	_, err = client.CoreV1().ConfigMaps(p.dataNamespace).Update(cm)

	return err
}

func (p *Plugin) loadManifest(client kubernetes.Interface, deployName string) (string, error) {
	name := p.getManifestConfigMapName(deployName)

	cm, err := client.CoreV1().ConfigMaps(p.dataNamespace).Get(name, meta.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("can't find data for deployment %s (should be stored in namespace %s): %s", deployName, p.dataNamespace, err)
		}
		return "", err
	}

	manifest := cm.Data["manifest"]
	if len(manifest) == 0 {
		return "", fmt.Errorf("no manifest found in data for deployment %s (stored in configmap %s/%s", deployName, p.dataNamespace, name)
	}

	return manifest, nil
}

func (p *Plugin) deleteManifest(client kubernetes.Interface, deployName string) error {
	name := p.getManifestConfigMapName(deployName)

	err := client.CoreV1().ConfigMaps(p.dataNamespace).Delete(name, &meta.DeleteOptions{})
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	return err
}
