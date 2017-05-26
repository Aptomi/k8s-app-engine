package slinga

import (
	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"
	"strings"
)

// HelmCodeExecutor is an executor that uses Helm for deployment of apps on kubernetes
type HelmCodeExecutor struct {
	Code *Code
}

func HelmName(str string) string {
	r := strings.NewReplacer("#", "-", "_", "-")
	return r.Replace(str)
}

const (
	//tillerHost = "tiller-deploy.kube-system.svc.cluster.local:44134"
	tillerHost = "kapp-demo-1:40666"
)

func newHelmClient() *helm.Client {
	return helm.NewClient(helm.Host(tillerHost))
}

func findHelmRelease(helmClient *helm.Client, name string) (bool, error) {
	resp, err := helmClient.ListReleases()
	if err != nil {
		return false, nil
	}

	for _, rel := range resp.Releases {
		if rel.Name == name {
			return true, nil
		}
	}

	return false, nil
}

//func preparePar

// Install for HelmCodeExecutor runs "helm install" for the corresponding helm chart
func (executor HelmCodeExecutor) Install(key string, codeMetadata map[string]string, codeParams interface{}) error {
	releaseName := strings.ToLower(HelmName(key))

	chartName := codeMetadata["chartName"]

	helmClient := newHelmClient()

	// TODO check err separately
	if exists, err := findHelmRelease(helmClient, releaseName); exists && err == nil {
		// TODO log that it's already installed
		// TODO update release just in case
		return nil
	}

	chartPath := GetAptomiPolicyDir() + "/charts/" + chartName + ".tgz"

	vals, err := yaml.Marshal(codeParams)
	if err != nil {
		return err
	}

	glog.Infof("Installing new Helm release '%s' of '%s' (path: %s) with params:\n%s", releaseName, chartName, chartPath, string(vals))

	// TODO is it good to reuse name?
	_ /*resp*/, err = helmClient.InstallRelease(chartPath, "aptomi", helm.ReleaseName(releaseName), helm.ValueOverrides(vals), helm.InstallReuseName(true))
	if err != nil {
		return err
	}
	return nil
}

// Update for HelmCodeExecutor runs "helm update" for the corresponding helm chart
func (executor HelmCodeExecutor) Update(key string, codeMetadata map[string]string, codeParams interface{}) error {
	// TODO merge with Install
	releaseName := strings.ToLower(HelmName(key))

	chartName := codeMetadata["chartName"]

	helmClient := newHelmClient()

	chartPath := GetAptomiPolicyDir() + "/charts/" + chartName + ".tgz"

	vals, err := yaml.Marshal(codeParams)
	if err != nil {
		return err
	}

	glog.Infof("Upgrading Helm release '%s' of '%s' (path: %s) with params:\n%s", releaseName, chartName, chartPath, string(vals))

	// TODO is it good to reuse name?
	_ /*resp*/, err = helmClient.UpdateRelease(releaseName, chartPath, helm.UpdateValueOverrides(vals))
	if err != nil {
		return err
	}

	return nil
}

// Destroy for HelmCodeExecutor runs "helm delete" for the corresponding helm chart
func (executor HelmCodeExecutor) Destroy(key string) error {
	releaseName := strings.ToLower(HelmName(key))

	helmClient := newHelmClient()

	glog.Infof("Deleting Helm release '%s'", releaseName)

	if _, err := helmClient.DeleteRelease(releaseName, helm.DeletePurge(true)); err != nil {
		return err
	}

	return nil
}
