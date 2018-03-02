package k8sraw

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/k8s"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Aptomi/aptomi/pkg/util/sync"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/helm/pkg/kube"
	"strings"
)

// Plugin represents Kubernetes Raw code plugin that supports deploying specified k8s objects into the cluster
type Plugin struct {
	once          sync.Init
	cluster       *lang.Cluster
	config        config.K8sRaw
	kube          *k8s.Plugin
	dataNamespace string
}

// New returns new instance of the Kubernetes Raw code (objects) plugin for specified Kubernetes cluster plugin and plugins config
func New(clusterPlugin plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
	kubePlugin, ok := clusterPlugin.(*k8s.Plugin)
	if !ok {
		return nil, fmt.Errorf("k8s cluster plugin expected for k8sraw code plugin creation but received: %T", clusterPlugin)
	}

	return &Plugin{
		cluster: kubePlugin.Cluster,
		config:  cfg.K8sRaw,
		kube:    kubePlugin,
	}, nil
}

func (p *Plugin) init() error {
	return p.once.Do(func() error {
		err := p.kube.Init()
		if err != nil {
			return err
		}

		err = p.parseClusterConfig()
		if err != nil {
			return err
		}

		kubeClient, err := p.kube.NewClient()
		if err != nil {
			return err
		}

		return p.kube.EnsureNamespace(kubeClient, p.dataNamespace)
	})
}

// Cleanup implements cleanup phase for the k8s raw plugin
func (p *Plugin) Cleanup() error {
	return nil
}

// Create implements creation of a new component instance in the cloud by deploying raw k8s objects
func (p *Plugin) Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.prepareClient(eventLog, deployName)

	err = client.Create(p.kube.Namespace, strings.NewReader(targetManifest), 42, false)
	if err != nil {
		return err
	}

	return p.storeManifest(kubeClient, deployName, targetManifest)
}

// Update implements update of an existing component instance in the cloud by updating raw k8s objects
func (p *Plugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	currentManifest, err := p.loadManifest(kubeClient, deployName)
	if err != nil {
		return err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.prepareClient(eventLog, deployName)

	err = client.Update(p.kube.Namespace, strings.NewReader(currentManifest), strings.NewReader(targetManifest), false, false, 42, false)
	if err != nil {
		return err
	}

	return p.storeManifest(kubeClient, deployName, targetManifest)
}

// Destroy implements destruction of an existing component instance in the cloud by deleting raw k8s objects
func (p *Plugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	err := p.init()
	if err != nil {
		return err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return err
	}

	deleteManifest, ok := params["manifest"].(string)
	if !ok {
		return fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.prepareClient(eventLog, deployName)

	err = client.Delete(p.kube.Namespace, strings.NewReader(deleteManifest))
	if err != nil {
		return err
	}

	return p.deleteManifest(kubeClient, deployName)
}

// Endpoints returns map from port type to url for all services of the deployed raw k8s objects
func (p *Plugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return nil, err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.prepareClient(eventLog, deployName)

	infos, err := client.BuildUnstructured(p.kube.Namespace, strings.NewReader(targetManifest))
	if err != nil {
		return nil, err
	}

	endpoints := make(map[string]string)

	for _, info := range infos {
		if info.Mapping.GroupVersionKind.Kind == "Service" {
			service, getErr := kubeClient.CoreV1().Services(p.kube.Namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return nil, getErr
			}

			p.kube.AddEndpointsFromService(service, endpoints)
		}
	}

	return endpoints, nil
}

func (p *Plugin) Status(deployName string, params util.NestedParameterMap, eventLog *event.Log) (plugin.DeploymentStatus, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	kubeClient, err := p.kube.NewClient()
	if err != nil {
		return nil, err
	}

	targetManifest, ok := params["manifest"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest is a mandatory parameter")
	}

	client := p.prepareClient(eventLog, deployName)

	infos, err := client.BuildUnstructured(p.kube.Namespace, strings.NewReader(targetManifest))
	if err != nil {
		return nil, err
	}

	handlers := make(map[string]ResourceTypeHandler)
	handlers["k8s/v1/Service"] = &K8sServiceResourceTypeHandler{}
	// not sure if it's good to have version.... we could have issues with versions in different k8s clusters
	handlers["k8s/v1/Deployment"] = &K8sDeploymentResourceTypeHandler{}

	status := make(plugin.DeploymentStatus)
	for _, info := range infos {
		gvk := info.ResourceMapping().GroupVersionKind
		resourceType := "k8s/" + gvk.Version + "/" + gvk.Kind

		handler, exist := handlers[resourceType]
		if !exist {
			continue
		}

		table, exist := status[resourceType]
		if !exist {
			table = &plugin.ResourceTable{}
			status[resourceType] = table
			table.Headers = handler.Headers()
		}

		if info.Mapping.GroupVersionKind.Kind == "Service" {
			service, getErr := kubeClient.CoreV1().Services(p.kube.Namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return nil, getErr
			}
			table.Items = append(table.Items, handler.Columns(service))
		} else if info.Mapping.GroupVersionKind.Kind == "Deployment" {
			deployment, getErr := kubeClient.AppsV1beta1().Deployments(p.kube.Namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return nil, getErr
			}
			table.Items = append(table.Items, handler.Columns(deployment))
		}
	}

	return status, nil
}

type ResourceTypeHandler interface {
	Headers() []string
	Columns(interface{}) []string
}

var K8sServiceResourceHeaders = []string{
	"Namespace",
	"Name",
	"Type",
	"Port(s)",
	"Created",
}

type K8sServiceResourceTypeHandler struct {
}

func (*K8sServiceResourceTypeHandler) Headers() []string {
	return K8sServiceResourceHeaders
}

func (*K8sServiceResourceTypeHandler) Columns(obj interface{}) []string {
	service := obj.(*v1.Service)
	parts := make([]string, len(service.Spec.Ports))
	for idx, port := range service.Spec.Ports {
		if port.NodePort > 0 {
			parts[idx] = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		} else {
			parts[idx] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		}
		if len(port.Name) > 0 {
			parts[idx] += "(" + port.Name + ")"
		}
	}
	ports := strings.Join(parts, ",")

	return []string{service.Namespace, service.Name, string(service.Spec.Type), ports, service.CreationTimestamp.String()}
}

var K8sDeploymentResourceHeaders = []string{
	"Namespace",
	"Name",
	"Desired",
	"Current",
	"Up-to-date",
	"Available",
	"Created",
}

type K8sDeploymentResourceTypeHandler struct {
}

func (*K8sDeploymentResourceTypeHandler) Headers() []string {
	return K8sDeploymentResourceHeaders
}

func (*K8sDeploymentResourceTypeHandler) Columns(obj interface{}) []string {
	deployment := obj.(*v1beta1.Deployment)

	desiredReplicas := fmt.Sprintf("%d", *deployment.Spec.Replicas)
	currentReplicas := fmt.Sprintf("%d", deployment.Status.Replicas)
	updatedReplicas := fmt.Sprintf("%d", deployment.Status.UpdatedReplicas)
	availableReplicas := fmt.Sprintf("%d", deployment.Status.AvailableReplicas)

	return []string{deployment.Namespace, deployment.Name, desiredReplicas, currentReplicas, updatedReplicas, availableReplicas, deployment.CreationTimestamp.String()}
}

func (p *Plugin) prepareClient(eventLog *event.Log, deployName string) *kube.Client {
	client := kube.New(p.kube.ClientConfig)
	client.Log = func(format string, args ...interface{}) {
		eventLog.WithFields(event.Fields{
			"deployName": deployName,
		}).Debugf(fmt.Sprintf("[instance: %s] ", deployName)+format, args...)
	}

	return client
}
