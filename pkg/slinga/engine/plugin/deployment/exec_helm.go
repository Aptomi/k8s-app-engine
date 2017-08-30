package deployment

import (
	"fmt"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/kubernetes/pkg/api"
	k8serrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	internalversioncore "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/labels"

	"strings"

	"errors"
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/mattn/go-zglob"
)

// HelmCodeExecutor is an executor that uses Helm for deployment of apps on kubernetes
type HelmCodeExecutor struct {
	Code     *Code
	Cluster  *Cluster
	Key      string
	Params   NestedParameterMap
	eventLog *EventLog
}

// NewHelmCodeExecutor constructs HelmCodeExecutor from given *Code
func NewHelmCodeExecutor(code *Code, key string, codeParams NestedParameterMap, clusters map[string]*Cluster, eventLog *EventLog) (CodeExecutor, error) {
	// First of all, redirect Helm/grpc logging to our event log
	// We don't want these messages to be printed to Stdout/Stderr
	grpclog.SetLogger(eventLog)

	// Get cluster name from code params
	clusterName, ok := codeParams["cluster"].(string)
	if !ok {
		return nil, errors.New("Cluster name not found in code params")
	}

	// Get cluster itself
	cluster, ok := clusters[clusterName]
	if !ok {
		return nil, errors.New("Cluster not found in policy: " + clusterName)
	}

	// Create code executor
	exec := HelmCodeExecutor{Code: code, Cluster: cluster, Key: key, Params: codeParams, eventLog: eventLog}
	err := exec.setupTillerConnection()
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func NewKubeClient(cluster *Cluster) (*restclient.Config, *internalclientset.Clientset, error) {
	kubeContext := cluster.Config.KubeContext
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

// HttpServices returns list of services for the current chart
func (exec HelmCodeExecutor) HttpServices() ([]string, error) {
	_, clientset, err := NewKubeClient(exec.Cluster)
	if err != nil {
		return nil, err
	}

	coreClient := clientset.Core()

	releaseName := releaseName(exec.Key)
	chartName, err := exec.chartName()
	if err != nil {
		return nil, err
	}

	selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector()
	options := api.ListOptions{LabelSelector: selector}

	// Check all corresponding services
	services, err := coreClient.Services(exec.Cluster.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	// Check all corresponding Istio ingresses
	ingresses, err := clientset.Extensions().Ingresses(exec.Cluster.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	if len(ingresses.Items) > 0 {
		result := make([]string, 0)
		for _, service := range services.Items {
			result = append(result, service.Name)
		}

		return result, nil
	}

	return nil, nil
}

func (exec *HelmCodeExecutor) setupTillerConnection() error {
	if len(exec.Cluster.GetTillerHost()) > 0 {
		// connection already set up, skip
		return nil
	}

	config, client, err := NewKubeClient(exec.Cluster)
	if err != nil {
		return err
	}

	tillerNamespace := exec.Cluster.Config.TillerNamespace
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	tunnel, err := portforwarder.New(tillerNamespace, client, config)
	if err != nil {
		return err
	}

	exec.Cluster.SetTillerHost(fmt.Sprintf("localhost:%d", tunnel.Local))

	exec.eventLog.WithFields(Fields{}).Debugf("Created k8s tunnel using local port: %s", tunnel.Local)

	return nil
}

func (exec *HelmCodeExecutor) newHelmClient() *helm.Client {
	return helm.NewClient(helm.Host(exec.Cluster.GetTillerHost()))
}

func findHelmRelease(helmClient *helm.Client, name string) (bool, error) {
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

func releaseName(key string) string {
	return strings.ToLower(EscapeName(key))
}

func (exec *HelmCodeExecutor) chartName() (string, error) {
	if chartName, ok := exec.Params["chartName"].(string); ok {
		return chartName, nil
	}

	return "", fmt.Errorf("Executor params don't contain chartName: key='%s', params='%s'", exec.Key, exec.Params)
}

// Install for HelmCodeExecutor runs "helm install" for the corresponding helm chart
func (exec HelmCodeExecutor) Install() error {
	releaseName := releaseName(exec.Key)
	chartName, err := exec.chartName()
	if err != nil {
		return err
	}

	helmClient := exec.newHelmClient()

	exists, err := findHelmRelease(helmClient, releaseName)
	if err != nil {
		return fmt.Errorf("Error while looking for Helm release '%s': %s", releaseName, err.Error())
	}

	if exists {
		// If a release already exists, let's just go ahead and update it
		exec.eventLog.WithFields(Fields{}).Infof("Release '%s' already exists. Updating it", releaseName)
		return exec.Update()
	}

	chartPath, err := getValidChartPath(chartName)
	if err != nil {
		return err
	}

	vals, err := yaml.Marshal(exec.Params)
	if err != nil {
		return err
	}

	exec.eventLog.WithFields(Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(vals),
	}).Infof("Installing Helm release '%s', chart '%s'", releaseName, chartName)

	// TODO: why is it always installing into a "demo" namespace?
	_, err = helmClient.InstallRelease(chartPath, "demo", helm.ReleaseName(releaseName), helm.ValueOverrides(vals), helm.InstallReuseName(true))
	return err
}

// Update for HelmCodeExecutor runs "helm update" for the corresponding helm chart
func (exec HelmCodeExecutor) Update() error {
	releaseName := releaseName(exec.Key)
	chartName, err := exec.chartName()
	if err != nil {
		return err
	}

	helmClient := exec.newHelmClient()

	chartPath, err := getValidChartPath(chartName)
	if err != nil {
		return err
	}

	vals, err := yaml.Marshal(exec.Params)
	if err != nil {
		return err
	}

	exec.eventLog.WithFields(Fields{
		"release": releaseName,
		"chart":   chartName,
		"path":    chartPath,
		"params":  string(vals),
	}).Infof("Updating Helm release '%s', chart '%s'", releaseName, chartName)

	_, err = helmClient.UpdateRelease(releaseName, chartPath, helm.UpdateValueOverrides(vals))
	return err
}

func getValidChartPath(chartName string) (string, error) {
	files, _ := zglob.Glob(GetAptomiObjectFilePatternTgz(GetAptomiBaseDir(), TypeCharts, chartName))
	fileName, err := EnsureSingleFile(files)
	if err != nil {
		return "", fmt.Errorf("Error while doing chart '%s' lookup: %s", chartName, err.Error())
	}
	return fileName, nil
}

// Destroy for HelmCodeExecutor runs "helm delete" for the corresponding helm chart
func (exec HelmCodeExecutor) Destroy() error {
	releaseName := releaseName(exec.Key)

	helmClient := exec.newHelmClient()

	exec.eventLog.WithFields(Fields{
		"release": releaseName,
	}).Infof("Deleting Helm release '%s'", releaseName)

	_, err := helmClient.DeleteRelease(releaseName, helm.DeletePurge(true))
	return err
}

// Endpoints returns map from port type to url for all services of the current chart
func (exec HelmCodeExecutor) Endpoints() (map[string]string, error) {
	_, clientset, err := NewKubeClient(exec.Cluster)
	if err != nil {
		return nil, err
	}

	coreClient := clientset.Core()

	releaseName := releaseName(exec.Key)
	chartName, err := exec.chartName()
	if err != nil {
		return nil, err
	}

	selector := labels.Set{"release": releaseName, "chart": chartName}.AsSelector()
	options := api.ListOptions{LabelSelector: selector}

	endpoints := make(map[string]string)

	// Check all corresponding services
	services, err := coreClient.Services(exec.Cluster.Config.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	kubeHost, err := exec.getKubeExternalAddress(coreClient)
	if err != nil {
		return nil, err
	}

	for _, service := range services.Items {
		if service.Spec.Type == "NodePort" {
			for _, port := range service.Spec.Ports {
				sURL := fmt.Sprintf("%s:%d", kubeHost, port.NodePort)

				// todo(slukjanov): could we somehow detect real schema? I think no :(
				if StringContainsAny(port.Name, "https") {
					sURL = "https://" + sURL
				} else if StringContainsAny(port.Name, "ui", "rest", "http", "grafana") {
					sURL = "http://" + sURL
				}

				endpoints[port.Name] = sURL
			}
		}
	}

	// Find Istio Ingress service (how ingress itself exposed)
	service, err := coreClient.Services(exec.Cluster.Config.Namespace).Get("istio-ingress")
	if err != nil {
		// return if there is no Istio deployed
		if k8serrors.IsNotFound(err) {
			return endpoints, nil
		}
		return nil, err
	}

	istioIngress := "<unresolved>"
	if service.Spec.Type == "NodePort" {
		for _, port := range service.Spec.Ports {
			if port.Name == "http" {
				istioIngress = fmt.Sprintf("%s:%d", kubeHost, port.NodePort)
			}
		}
	}

	// Check all corresponding istio ingresses
	ingresses, err := clientset.Extensions().Ingresses(exec.Cluster.Config.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	// todo(slukjanov): support more then one ingress / rule / path
	for _, ingress := range ingresses.Items {
		if class, ok := ingress.Annotations["kubernetes.io/ingress.class"]; !ok || class != "istio" {
			continue
		}
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				pathStr := strings.Trim(path.Path, ".*")

				if rule.Host == "" {
					endpoints["ingress"] = "http://" + istioIngress + pathStr
				} else {
					endpoints["ingress"] = "http://" + rule.Host + pathStr
				}
			}
		}
	}

	return endpoints, nil
}

func (exec HelmCodeExecutor) getKubeExternalAddress(client internalversioncore.CoreInterface) (string, error) {
	if len(exec.Cluster.GetKubeExternalAddress()) > 0 {
		return exec.Cluster.GetKubeExternalAddress(), nil
	}

	nodes, err := client.Nodes().List(api.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", errors.New("K8s nodes list if empty, fail")
	}

	returnFirst := func(addrType api.NodeAddressType) string {
		for _, node := range nodes.Items {
			for _, addr := range node.Status.Addresses {
				if addr.Type == addrType {
					return addr.Address
				}
			}
		}
		return ""
	}

	addr := returnFirst(api.NodeExternalIP)
	if addr == "" {
		addr = returnFirst(api.NodeLegacyHostIP)
	}
	if addr == "" {
		addr = returnFirst(api.NodeInternalIP)
	}
	if addr == "" {
		return "", errors.New("Couldn't find external IP for cluster")
	}

	exec.Cluster.SetKubeExternalAddress(addr)
	return addr, nil
}
