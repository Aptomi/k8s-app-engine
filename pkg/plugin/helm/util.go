package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/util"
	"io/ioutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/repo"
	"strings"
)

func (plugin *Plugin) newClient() (*helm.Client, error) {
	return helm.NewClient(helm.Host(plugin.tillerHost)), nil
}

func getHelmReleaseInfo(params util.NestedParameterMap) (repository, name, version string, err error) {
	var ok bool
	if repository, ok = params["chartRepo"].(string); !ok {
		err = fmt.Errorf("chartRepo is a mandatory parameter")
		return
	}

	if name, ok = params["chartName"].(string); !ok {
		err = fmt.Errorf("chartName is a mandatory parameter")
		return
	}

	if _, ok = params["chartVersion"]; !ok {
		// version is optional. this will use the latest
		version = ""
	} else {
		version = params["chartVersion"].(string)
	}

	return
}

var (
	releaseNameReplacer = strings.NewReplacer("#", "-", "_", "-")
)

func getReleaseName(deployName string) string {
	return strings.ToLower(releaseNameReplacer.Replace(deployName))
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
