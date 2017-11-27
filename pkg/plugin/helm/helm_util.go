package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	"io/ioutil"
	"k8s.io/helm/cmd/helm/installer"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/repo"
	"strings"
)

func getHelmReleaseInfo(params util.NestedParameterMap) (repository, name, version string, err error) {
	repository, ok := params["chartRepo"].(string)
	if !ok {
		err = fmt.Errorf("chartRepo is a mandatory paraneter")
		return
	}

	name, ok = params["chartName"].(string)
	if !ok {
		err = fmt.Errorf("chartName is a mandatory paraneter")
		return
	}

	version = params["chartVersion"].(string)

	return
}

func getHelmReleaseName(deployName string) string {
	return strings.ToLower(util.EscapeName(deployName))
}

func findHelmRelease(helmClient *helm.Client, name string) (bool, error) {
	// todo(slukjanov): use release list filter
	resp, err := helmClient.ListReleases()
	if err != nil {
		return false, err
	}

	for _, release := range resp.Releases {
		if release.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (cache *clusterCache) newHelmClient(eventLog *event.Log) (*helm.Client, error) {
	return helm.NewClient(helm.Host(cache.tillerHost)), nil
}

func (cache *clusterCache) ensureTillerTunnel(eventLog *event.Log) error {
	if len(cache.tillerHost) > 0 {
		// todo(slukjanov): verify that tunnel is still alive??
		// connection already set up, skip
		return nil
	}

	var tunnelErr error
	ok := retry.Do(120, 5, func() bool {
		cache.tillerTunnel, tunnelErr = cache.newTillerTunnel()

		if tunnelErr != nil {
			if strings.Contains(tunnelErr.Error(), "could not find tiller") {
				err := cache.setupTiller(eventLog)
				if err != nil {
					tunnelErr = err
					return false
				}
			}

			return false
		}

		port := cache.tillerTunnel.Local
		cache.tillerHost = fmt.Sprintf("localhost:%d", port)
		eventLog.WithFields(event.Fields{}).Debugf("Created k8s tunnel using local port: %d", port)

		return true
	})

	if !ok {
		if tunnelErr != nil {
			return tunnelErr
		}

		return fmt.Errorf("tiller tunnel creation timeout for cluster: %s", cache.cluster.Name)
	}

	return nil
}

func (cache *clusterCache) newTillerTunnel() (*kube.Tunnel, error) {
	client, err := cache.newKubeClient()
	if err != nil {
		return nil, err
	}

	return portforwarder.New(cache.tillerNamespace, client, cache.kubeConfig)
}

func (cache *clusterCache) setupTiller(eventLog *event.Log) error {
	client, err := cache.newKubeClient()
	if err != nil {
		return err
	}

	eventLog.WithFields(event.Fields{}).Debugf("Setting up tiller in cluster %s namespace %s", cache.cluster.Name, cache.tillerNamespace)

	err = cache.ensureKubeNamespace(client, cache.tillerNamespace)
	if err != nil {
		return err
	}

	err = cache.createKubeServiceAccount(client, cache.tillerNamespace)
	if err != nil {
		return err
	}

	err = cache.createKubeClusterRoleBinding(client, cache.tillerNamespace)
	if err != nil {
		return err
	}

	return installer.Install(client, &installer.Options{
		Namespace:      cache.tillerNamespace,
		ImageSpec:      "gcr.io/kubernetes-helm/tiller:v2.6.2",
		ServiceAccount: "tiller-" + cache.tillerNamespace,
	})
}

func (plugin *Plugin) fetchChart(repository, name, version string) (string, error) {
	chartURL, err := repo.FindChartInRepoURL(
		repository, name, version,
		"", "", "",
		newGetterProviders(),
	)
	if err != nil {
		return "", fmt.Errorf("error while getting chart url in repo: %s", err)
	}

	chartFile, err := ioutil.TempFile("", name)
	if err != nil {
		return "", fmt.Errorf("error while creating temp file for downloading chart: %s", err)
	}

	chartGetter, err := newHTTPGetter(chartURL, "", "", "")
	if err != nil {
		return "", fmt.Errorf("error while creating chart downloader: %s", err)
	}

	resp, err := chartGetter.Get(chartURL)
	if err != nil {
		return "", fmt.Errorf("error while downloading chart: %s", err)
	}

	_, err = chartFile.Write(resp.Bytes())
	if err != nil {
		return "", fmt.Errorf("error while writing downloaded chart to the temp file")
	}

	return chartFile.Name(), nil
}
