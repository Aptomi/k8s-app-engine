package slinga

import (
	"github.com/golang/glog"
	"k8s.io/helm/pkg/helm"
	"strings"
	yaml "gopkg.in/yaml.v2"
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

// Install for HelmCodeExecutor runs "helm install" for the corresponding helm chart
func (executor HelmCodeExecutor) Install(key string, codeMetadata map[string]string, codeParams interface{}) error {
	uid := strings.ToLower(HelmName(key))

	chartName := codeMetadata["chartName"]

	helmClient := newHelmClient()

	// TODO check err separately
	if exists, err := findHelmRelease(helmClient, uid); exists && err == nil {
		// TODO log that it's already installed
		// TODO update release just in case
		return nil
	}

	chartPath := GetAptomiPolicyDir() + "/charts/" + chartName + ".tgz"

	vals, err := yaml.Marshal(codeParams)
	if err != nil {
		return err
	}

	//TODO: print how we're running helm
	glog.Info("Running Helm install for: ", uid, " ", chartPath, "\n", string(vals))

	// TODO is it good to reuse name?
	_ /*resp*/, err = helmClient.InstallRelease(chartPath, "aptomi", helm.ReleaseName(uid), helm.ValueOverrides(vals), helm.InstallReuseName(true))
	if err != nil {
		return err
	}
	return nil
}

// Update for HelmCodeExecutor runs "helm update" for the corresponding helm chart
func (executor HelmCodeExecutor) Update(key string, labels LabelSet) error {
	// TODO: implement update method
	return nil
}

// Destroy for HelmCodeExecutor runs "helm delete" for the corresponding helm chart
func (executor HelmCodeExecutor) Destroy(key string) error {
	uid := strings.ToLower(HelmName(key))

	helmClient := newHelmClient()
	if _, err := helmClient.DeleteRelease(uid); err != nil {
		return err
	}

	return nil
}
