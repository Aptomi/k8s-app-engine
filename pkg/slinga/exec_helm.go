package slinga

import (
	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"
	"strings"
)

// HelmCodeExecutor is an executor that uses Helm for deployment of apps on kubernetes
type HelmCodeExecutor struct {
	Code *Code
	tillerHost string
}

// NewHelmCodeExecutor constructs HelmCodeExecutor from given *Code
func NewHelmCodeExecutor(code *Code, tillerHost string) CodeExecutor {
	// First of all, redirect Helm/grpc logging to our own debug stream
	// We don't want these messages to be printed to Stdout/Stderr
	grpclog.SetLogger(debug)

	// Next, create the executor itself
	return HelmCodeExecutor{Code: code, tillerHost: tillerHost}
}

func (executor *HelmCodeExecutor) newHelmClient() *helm.Client {
	return helm.NewClient(helm.Host(executor.tillerHost))
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

func releaseName( key string) string {
	return strings.ToLower(EscapeName(key))
}

// Install for HelmCodeExecutor runs "helm install" for the corresponding helm chart
func (executor HelmCodeExecutor) Install(key string, codeMetadata map[string]string, codeParams interface{}) error {
	releaseName := releaseName(key)

	chartName := codeMetadata["chartName"]

	helmClient := executor.newHelmClient()

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

	debug.WithFields(log.Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(vals),
	}).Info("Installing Helm release")

	_, err = helmClient.InstallRelease(chartPath, "demo", helm.ReleaseName(releaseName), helm.ValueOverrides(vals), helm.InstallReuseName(true))
	if err != nil {
		return err
	}
	return nil
}

// Update for HelmCodeExecutor runs "helm update" for the corresponding helm chart
func (executor HelmCodeExecutor) Update(key string, codeMetadata map[string]string, codeParams interface{}) error {
	releaseName := releaseName(key)

	chartName := codeMetadata["chartName"]

	helmClient := executor.newHelmClient()

	chartPath := GetAptomiPolicyDir() + "/charts/" + chartName + ".tgz"

	vals, err := yaml.Marshal(codeParams)
	if err != nil {
		return err
	}

	debug.WithFields(log.Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(vals),
	}).Info("Updating Helm release")

	_, err = helmClient.UpdateRelease(releaseName, chartPath, helm.UpdateValueOverrides(vals))
	if err != nil {
		return err
	}

	return nil
}

// Destroy for HelmCodeExecutor runs "helm delete" for the corresponding helm chart
func (executor HelmCodeExecutor) Destroy(key string) error {
	releaseName := releaseName(key)

	helmClient := executor.newHelmClient()

	debug.WithFields(log.Fields{
		"release": releaseName,
	}).Info("Deleting Helm release")

	if _, err := helmClient.DeleteRelease(releaseName, helm.DeletePurge(true)); err != nil {
		return err
	}

	return nil
}
