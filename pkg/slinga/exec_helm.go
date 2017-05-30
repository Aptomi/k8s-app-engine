package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/labels"
	"strings"

	"errors"
	"net"
	"net/url"
)

// HelmCodeExecutor is an executor that uses Helm for deployment of apps on kubernetes
type HelmCodeExecutor struct {
	Code     *Code
	Cluster  *Cluster
	Key      string
	Metadata map[string]string
	Params   interface{}
}

// NewHelmCodeExecutor constructs HelmCodeExecutor from given *Code
func NewHelmCodeExecutor(code *Code, key string, codeMetadata map[string]string, codeParams interface{}, clusters map[string]*Cluster) (CodeExecutor, error) {
	// First of all, redirect Helm/grpc logging to our own debug stream
	// We don't want these messages to be printed to Stdout/Stderr
	grpclog.SetLogger(debug)

	if paramsMap, ok := codeParams.(map[interface{}]interface{}); ok {
		// todo: should we check key existence first?
		if clusterName, ok := paramsMap["cluster"].(string); !ok {
			return nil, errors.New("Cluster param should be defined")
		} else if cluster, ok := clusters[clusterName]; ok {
			exec := HelmCodeExecutor{Code: code, Cluster: cluster, Key: key, Metadata: codeMetadata, Params: codeParams}
			err := exec.setupTillerConnection()
			if err != nil {
				return nil, err
			}
			return exec, nil
		} else {
			return nil, errors.New("Specified cluster is undefined")
		}
	}
	return nil, errors.New("Can't parse codeParams")
}

func (exec *HelmCodeExecutor) newKubeClient() (*restclient.Config, *internalclientset.Clientset, error) {
	kubeContext := exec.Cluster.Metadata.KubeContext
	config, err := kube.GetConfig(kubeContext).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get kubernetes config for context '%s': %s", kubeContext, err)
	}
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get kubernetes client: %s", err)
	}
	return config, client, nil
}

func (exec *HelmCodeExecutor) setupTillerConnection() error {
	if exec.Cluster.Metadata.tillerHost != "" {
		return nil
	}

	config, client, err := exec.newKubeClient()
	if err != nil {
		return err
	}

	tillerNamespace := exec.Cluster.Metadata.TillerNamespace
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	tunnel, err := portforwarder.New(tillerNamespace, client, config)
	if err != nil {
		return err
	}

	exec.Cluster.Metadata.tillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)

	debug.WithFields(log.Fields{
		"port": tunnel.Local,
	}).Info("Created k8s tunnel using local port")

	return nil
}

func (exec *HelmCodeExecutor) newHelmClient() *helm.Client {
	return helm.NewClient(helm.Host(exec.Cluster.Metadata.tillerHost))
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

func releaseName(key string) string {
	return strings.ToLower(EscapeName(key))
}

// Install for HelmCodeExecutor runs "helm install" for the corresponding helm chart
func (exec HelmCodeExecutor) Install() error {
	releaseName := releaseName(exec.Key)

	chartName, ok := exec.Metadata["chartName"]
	if !ok {
		return errors.New("Chart name is undefined")
	}

	helmClient := exec.newHelmClient()

	exists, err := findHelmRelease(helmClient, releaseName)
	if err != nil {
		debug.WithFields(log.Fields{
			"releaseName": releaseName,
			"error":       err,
		}).Fatal("Err while looking for release")
	}

	if exists {
		debug.WithFields(log.Fields{
			"releaseName": releaseName,
		}).Fatal("Release already exists")
	}

	chartPath := GetAptomiPolicyDir() + "/charts/" + chartName + ".tgz"

	vals, err := yaml.Marshal(exec.Params)
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
func (exec HelmCodeExecutor) Update() error {
	releaseName := releaseName(exec.Key)

	chartName := exec.Metadata["chartName"]

	helmClient := exec.newHelmClient()

	chartPath := GetAptomiPolicyDir() + "/charts/" + chartName + ".tgz"

	vals, err := yaml.Marshal(exec.Params)
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
func (exec HelmCodeExecutor) Destroy() error {
	releaseName := releaseName(exec.Key)

	helmClient := exec.newHelmClient()

	debug.WithFields(log.Fields{
		"release": releaseName,
	}).Info("Deleting Helm release")

	if _, err := helmClient.DeleteRelease(releaseName, helm.DeletePurge(true)); err != nil {
		return err
	}

	return nil
}

// Endpoints returns map from port type to url for all services of the current chart
func (exec HelmCodeExecutor) Endpoints() (map[string]string, error) {
	_, client, err := exec.newKubeClient()
	if err != nil {
		return nil, err
	}

	if svcGetter, ok := client.Core().(internalversion.ServicesGetter); ok {
		releaseName := releaseName(exec.Key)
		chartName := exec.Metadata["chartName"]

		selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector()
		options := api.ListOptions{LabelSelector: selector}
		services, err := svcGetter.Services(exec.Cluster.Metadata.Namespace).List(options)
		if err != nil {
			return nil, err
		}
		endpoints := make(map[string]string)

		kubeHost, err := exec.getKubeHost()
		if err != nil {
			return nil, err
		}

		for _, service := range services.Items {
			if service.Spec.Type == "NodePort" {
				for _, port := range service.Spec.Ports {
					sURL := fmt.Sprintf("%s:%d", kubeHost, port.NodePort)

					// todo(slukjanov): could we somehow detect real schema? I think no :(
					if port.Name == "webui" || port.Name == "ui" || port.Name == "rest" {
						sURL = "http://" + sURL
					}

					endpoints[port.Name] = sURL
				}
			}
		}
		return endpoints, nil
	}

	return nil, nil
}

func (exec HelmCodeExecutor) getKubeHost() (string, error) {
	config, _, err := exec.newKubeClient()
	if err != nil {
		return "", err
	}

	u, err := url.Parse(config.Host)
	if err != nil {
		return "", err
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", err
	}

	return host, nil
}
