package k8s

import (
	"github.com/Aptomi/aptomi/pkg/event"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/kubectl"
	"strings"
	"k8s.io/client-go/kubernetes"
)

// ReadinessStatusForManifest returns readiness status of all resources for specified manifest
func (p *Plugin) ReadinessStatusForManifest(namespace, deployName, targetManifest string, eventLog *event.Log) (bool, error) {
	kubeClient, err := p.NewClient()
	if err != nil {
		return false, err
	}

	helmKube := p.NewHelmKube(deployName, eventLog)

	infos, err := helmKube.BuildUnstructured(namespace, strings.NewReader(targetManifest))
	if err != nil {
		return false, err
	}

	ready := true

	internalClientSet, clientErr := kubernetes.NewForConfig(p.RestConfig)
	if clientErr != nil {
		return false, clientErr
	}

	for _, info := range infos {
		if !ready {
			return false, nil
		}

		var statusErr error

		// todo some objects are missing in this check like DaemonSet, Job, ReplicationController, etc.
		switch kind := info.Mapping.GroupVersionKind.Kind; kind {
		case "Service": // nolint: goconst
			svc, getErr := kubeClient.CoreV1().Services(namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return false, getErr
			}
			ready = isServiceReady(svc)
		case "PersistentVolumeClaim":
			pvc, getErr := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return false, getErr
			}
			ready = isPersistentVolumeClaimReady(pvc)
		case "Deployment":
			//deployment, getErr := kubeClient.AppsV1beta1().Deployments(p.Namespace).Get(info.Name, meta.GetOptions{})
			//if getErr != nil {
			//	return false, getErr
			//}
			ready, statusErr = isReadyUsingStatusViewer(internalClientSet, apps.Kind("Deployment"), info.Namespace, info.Name)
		case "StatefulSet":
			//statefulSet, getErr := kubeClient.AppsV1beta1().StatefulSets(p.Namespace).Get(info.Name, meta.GetOptions{})
			//if getErr != nil {
			//	return false, getErr
			//}
			ready, statusErr = isReadyUsingStatusViewer(internalClientSet, apps.Kind("StatefulSet"), info.Namespace, info.Name)
		}

		if statusErr != nil {
			return false, nil
		}
	}

	return ready, nil
}

func isServiceReady(svc *v1.Service) bool {
	// skip checking external services
	if svc.Spec.Type == v1.ServiceTypeExternalName {
		return true
	}

	// check if cluster IP assigned if not explicitly set to none
	if svc.Spec.ClusterIP == "" {
		return false
	}

	// check address of load balancer type service
	if svc.Spec.Type == v1.ServiceTypeLoadBalancer && svc.Status.LoadBalancer.Ingress == nil {
		return false
	}

	return true
}

func isPersistentVolumeClaimReady(pvc *v1.PersistentVolumeClaim) bool {
	return pvc.Status.Phase == v1.ClaimBound
}

func isReadyUsingStatusViewer(internalClientSet kubernetes.Interface, groupKind schema.GroupKind, namespace, name string) (bool, error) {
	statusViewer, err := kubectl.StatusViewerFor(groupKind, internalClientSet)
	if err != nil {
		return false, err
	}

	_, ready, err := statusViewer.Status(namespace, name, 0)
	if err != nil {
		return false, err
	}

	return ready, nil
}
