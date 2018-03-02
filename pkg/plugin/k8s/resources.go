package k8s

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/plugin"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"strings"
)

var resourceRegistry = buildResourceRegistry()

func (p *Plugin) ResourcesForManifest(deployName, targetManifest string, eventLog *event.Log) (plugin.Resources, error) {
	kubeClient, err := p.NewClient()
	if err != nil {
		return nil, err
	}

	helmKube := p.NewHelmKube(deployName, eventLog)

	infos, err := helmKube.BuildUnstructured(p.Namespace, strings.NewReader(targetManifest))
	if err != nil {
		return nil, err
	}

	resources := make(plugin.Resources)
	for _, info := range infos {
		gvk := info.ResourceMapping().GroupVersionKind
		resourceType := "k8s/" + gvk.Version + "/" + gvk.Kind

		if !resourceRegistry.IsSupported(resourceType) {
			continue
		}

		table, exist := resources[resourceType]
		if !exist {
			table = &plugin.ResourceTable{}
			resources[resourceType] = table
			table.Headers = resourceRegistry.Headers(resourceType)
		}

		var getErr error
		var obj interface{}

		switch kind := info.Mapping.GroupVersionKind.Kind; kind {
		case "Service":
			obj, getErr = kubeClient.CoreV1().Services(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "ConfigMap":
			obj, getErr = kubeClient.CoreV1().ConfigMaps(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "Secret":
			obj, getErr = kubeClient.CoreV1().Secrets(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "PersistentVolumeClaim":
			obj, getErr = kubeClient.CoreV1().PersistentVolumeClaims(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "Deployment":
			obj, getErr = kubeClient.AppsV1beta1().Deployments(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "StatefulSet":
			obj, getErr = kubeClient.AppsV1beta1().StatefulSets(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "Job":
			obj, getErr = kubeClient.BatchV1().Jobs(p.Namespace).Get(info.Name, meta.GetOptions{})
		}

		if getErr != nil {
			return nil, getErr
		}
		table.Items = append(table.Items, resourceRegistry.Handle(resourceType, obj))
	}

	return resources, nil
}

func buildResourceRegistry() *plugin.ResourceRegistry {
	reg := plugin.NewResourceRegistry()

	reg.AddHandler("k8s/v1/Service", serviceResourceHeaders, serviceResourceHandler)
	reg.AddHandler("k8s/v1/Deployment", deploymentResourceHeaders, deploymentResourceHandler)

	return reg
}

// K8s Service

var serviceResourceHeaders = []string{
	"Namespace",
	"Name",
	"Type",
	"Port(s)",
	"Created",
}

func serviceResourceHandler(obj interface{}) []string {
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

// K8s Deployment

var deploymentResourceHeaders = []string{
	"Namespace",
	"Name",
	"Desired",
	"Current",
	"Up-to-date",
	"Available",
	"Generation",
	"Created",
}

func deploymentResourceHandler(obj interface{}) []string {
	deployment := obj.(*v1beta1.Deployment)

	desiredReplicas := fmt.Sprintf("%d", *deployment.Spec.Replicas)
	currentReplicas := fmt.Sprintf("%d", deployment.Status.Replicas)
	updatedReplicas := fmt.Sprintf("%d", deployment.Status.UpdatedReplicas)
	availableReplicas := fmt.Sprintf("%d", deployment.Status.AvailableReplicas)
	gen := fmt.Sprintf("%d", deployment.Generation)
	created := deployment.CreationTimestamp.String()

	return []string{deployment.Namespace, deployment.Name, desiredReplicas, currentReplicas, updatedReplicas, availableReplicas, gen, created}
}
