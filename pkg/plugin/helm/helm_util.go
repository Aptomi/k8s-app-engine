package helm

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/mattn/go-zglob"
	"k8s.io/helm/pkg/helm"
	"path/filepath"
	"strings"
)

func (cache *clusterCache) newHelmClient(cluster *lang.Cluster) *helm.Client {
	return helm.NewClient(helm.Host(cache.tillerHost))
}

func (p *Plugin) getValidChartPath(chartName string) (string, error) {
	pattern := filepath.Join(p.cfg.ChartsDir, "**", chartName+".tgz")
	files, err := zglob.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("error while searching chart %s file: %s", chartName, err)
	}
	fileName, err := util.EnsureSingleFile(files)
	if err != nil {
		return "", fmt.Errorf("Error while doing chart '%s' lookup: %s", chartName, err)
	}
	return fileName, nil
}

func helmChartName(params util.NestedParameterMap) (string, error) {
	if chartName, ok := params["chartName"].(string); ok {
		return chartName, nil
	}

	return "", fmt.Errorf("No chartName in params")
}

func helmReleaseName(deployName string) string {
	return strings.ToLower(util.EscapeName(deployName))
}

func findHelmRelease(helmClient *helm.Client, name string) (bool, error) {
	// todo(slukjanov): use release list filter
	resp, err := helmClient.ListReleases()
	if err != nil {
		return false, err
	}

	for _, rel := range resp.Releases {
		if rel.Name == name {
			return true, nil
		}
	}

	return false, nil
}
