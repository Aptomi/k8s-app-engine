package k8s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var resourceRegistry = buildResourceRegistry()

// ResourcesForManifest returns resources for specified manifest
func (p *Plugin) ResourcesForManifest(namespace, deployName, targetManifest string, eventLog *event.Log) (plugin.Resources, error) {
	kubeClient, err := p.NewClient()
	if err != nil {
		return nil, err
	}

	helmKube := p.NewHelmKube(deployName, eventLog)

	infos, err := helmKube.BuildUnstructured(namespace, strings.NewReader(targetManifest))
	if err != nil {
		return nil, err
	}

	resources := make(plugin.Resources)
	for _, info := range infos {
		gvk := info.ResourceMapping().GroupVersionKind
		resourceType := "k8s/" + gvk.Kind

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
		case "Service": // nolint: goconst
			obj, getErr = kubeClient.CoreV1().Services(namespace).Get(info.Name, meta.GetOptions{})
		case "ConfigMap":
			//obj, getErr = kubeClient.CoreV1().ConfigMaps(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "Secret":
			//obj, getErr = kubeClient.CoreV1().Secrets(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "PersistentVolumeClaim":
			//obj, getErr = kubeClient.CoreV1().PersistentVolumeClaims(p.Namespace).Get(info.Name, meta.GetOptions{})
		case "Deployment":
			obj, getErr = kubeClient.AppsV1beta1().Deployments(namespace).Get(info.Name, meta.GetOptions{})
		case "StatefulSet":
			obj, getErr = kubeClient.AppsV1beta1().StatefulSets(namespace).Get(info.Name, meta.GetOptions{})
		case "Job":
			//obj, getErr = kubeClient.BatchV1().Jobs(p.Namespace).Get(info.Name, meta.GetOptions{})
		}

		if getErr != nil {
			return nil, getErr
		}

		// missing object or don't know how to load it
		if obj == nil {
			continue
		}

		table.Items = append(table.Items, resourceRegistry.Handle(resourceType, obj))
	}

	return resources, nil
}

func buildResourceRegistry() *plugin.ResourceRegistry {
	reg := plugin.NewResourceRegistry()

	// todo(slukjanov): support all other resources
	// k8s/Service handling is temporarily disabled
	reg.AddHandler("_k8s/Service", serviceResourceHeaders, serviceResourceHandler)
	reg.AddHandler("k8s/Deployment", deploymentResourceHeaders, deploymentResourceHandler)
	reg.AddHandler("k8s/StatefulSet", statefulSetResourceHeaders, statefulSetResourceHandler)

	return reg
}

// k8s/Service

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

// k8s/Deployment

var deploymentResourceHeaders = []string{
	"Namespace",
	"Name",
	"Ready",
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
	ready := strconv.FormatBool(desiredReplicas == currentReplicas && currentReplicas == updatedReplicas && updatedReplicas == availableReplicas)
	gen := fmt.Sprintf("%d", deployment.Generation)
	created := deployment.CreationTimestamp.String()

	return []string{deployment.Namespace, deployment.Name, ready, desiredReplicas, currentReplicas, updatedReplicas, availableReplicas, gen, created}
}

// k8s/StatefulSet

var statefulSetResourceHeaders = []string{
	"Namespace",
	"Name",
	"Ready",
	"Desired",
	"Current",
}

func statefulSetResourceHandler(obj interface{}) []string {
	statefulSet := obj.(*v1beta1.StatefulSet)

	desiredReplicas := fmt.Sprintf("%d", *statefulSet.Spec.Replicas)
	currentReplicas := fmt.Sprintf("%d", statefulSet.Status.Replicas)
	ready := strconv.FormatBool(desiredReplicas == currentReplicas)

	return []string{statefulSet.Namespace, statefulSet.Name, ready, desiredReplicas, currentReplicas}
}
