package helm_istio

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/mattn/go-zglob"
	"k8s.io/helm/pkg/helm"
	"strings"
)

func (cache *clusterCache) newHelmClient(cluster *lang.Cluster) *helm.Client {
	return helm.NewClient(helm.Host(cache.tillerHost))
}

func helmChartName(params util.NestedParameterMap) (string, error) {
	if chartName, ok := params["chartName"].(string); ok {
		return chartName, nil
	}

	return "", fmt.Errorf("No chartName in params")
}

func getValidChartPath(chartName string) (string, error) {
	files, _ := zglob.Glob(db.GetAptomiObjectFilePatternTgz(db.GetAptomiBaseDir(), db.TypeCharts, chartName))
	fileName, err := util.EnsureSingleFile(files)
	if err != nil {
		return "", fmt.Errorf("Error while doing chart '%s' lookup: %s", chartName, err)
	}
	return fileName, nil
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
