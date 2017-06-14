package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

import (
	"k8s.io/kubernetes/pkg/api"
	k8slabels "k8s.io/kubernetes/pkg/labels"
)

// ProcessIstioIngress processes global rules and applies Istio routing rules for ingresses
func (usage *ServiceUsageState) ProcessIstioIngress(noop bool) {
	if len(usage.getResolvedUsage().ComponentProcessingOrder) == 0 || noop {
		return
	}

	fmt.Println("[Routes]")

	progress := NewProgress()
	progressBar := AddProgressBar(progress, len(usage.getResolvedUsage().ComponentProcessingOrder))

	desiredBlockedServices := make([]string, 0)

	// Process in the right order
	for _, key := range usage.getResolvedUsage().ComponentProcessingOrder {
		services, err := processComponent(key, usage)
		if err != nil {
			debug.WithFields(log.Fields{
				"key":   key,
				"error": err,
			}).Fatal("Unable to process Istio Ingress for component")
		}
		desiredBlockedServices = append(desiredBlockedServices, services...)
		progressBar.Incr()
	}

	progress.Stop()

	fmt.Println("Desired blocked services:", desiredBlockedServices)

	// todo(slukjanov): add actual istio route rules creation/deletion here
}

func processComponent(key string, usage *ServiceUsageState) ([]string, error) {
	serviceName, _, _, componentName := ParseServiceUsageKey(key)
	component := usage.Policy.Services[serviceName].getComponentsMap()[componentName]

	labels := usage.ResolvedUsage.ComponentInstanceMap[key].CalculatedLabels

	// todo(slukjanov): temp hack - expecting that cluster is always passed through the label "cluster"
	var cluster *Cluster
	if clusterLabel, ok := labels.Labels["cluster"]; ok {
		if cluster, ok = usage.Policy.Clusters[clusterLabel]; !ok {
			debug.WithFields(log.Fields{
				"component": key,
				"labels":    labels.Labels,
			}).Fatal("Can't find cluster for component (based on label 'cluster')")
		}
	}

	// get all users who're using service
	userIds := usage.ResolvedUsage.ComponentInstanceMap[key].UserIds
	users := make([]*User, 0)
	for _, userID := range userIds {
		// todo check if user doesn't exists
		users = append(users, usage.users.Users[userID])
	}

	if !usage.Policy.Rules.allowsIngressAccess(labels, users, cluster) && component != nil && component.Code != nil {
		codeExecutor, err := component.Code.GetCodeExecutor(key, component.Code.Metadata, usage.getResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, usage.Policy.Clusters)
		if err != nil {
			return nil, err
		}

		if helmCodeExecutor, ok := codeExecutor.(HelmCodeExecutor); ok {
			services, err := helmCodeExecutor.httpServices()
			if err != nil {
				return nil, err
			}

			for _, service := range services {
				content := "type: route-rule\n"
				content += "name: block-" + service + "\n"
				content += "spec:\n"
				content += "  destination: " + service + "." + cluster.Metadata.Namespace + ".svc.cluster.local\n"
				content += "  httpReqTimeout:\n"
				content += "    simpleTimeout:\n"
				content += "      timeout: 1ms\n"

				ruleFile := writeTempFile("istio-rule", content)
				fmt.Println("Istio rule file:", ruleFile.Name())
				//TODO: slukjanov: defer os.Remove(tmpFile.Name())

				content = "set -ex\n"
				content += "kubectl config use-context " + cluster.Name + "\n"
				// todo(slukjanov): find istio pilot service automatically
				content += "istioctl --configAPIService istio-prod-production-istio-istio-pilot:8081 --namespace " + cluster.Metadata.Namespace + " "
				content += "create -f " + ruleFile.Name() + "\n"

				cmdFile := writeTempFile("istioctl", content)
				fmt.Println("Istio cmd file:", cmdFile.Name())
			}

			return services, nil
		}
	}

	return nil, nil
}

// httpServices returns list of services for the current chart
func (exec HelmCodeExecutor) httpServices() ([]string, error) {
	_, clientset, err := exec.newKubeClient()
	if err != nil {
		return nil, err
	}

	coreClient := clientset.Core()

	releaseName := releaseName(exec.Key)
	chartName := exec.Metadata["chartName"]

	selector := k8slabels.Set{"release": releaseName, "chart": chartName}.AsSelector()
	options := api.ListOptions{LabelSelector: selector}

	// Check all corresponding services
	services, err := coreClient.Services(exec.Cluster.Metadata.Namespace).List(options)
	if err != nil {
		return nil, err
	}

	// Check all corresponding Istio ingresses
	ingresses, err := clientset.Extensions().Ingresses(exec.Cluster.Metadata.Namespace).List(options)
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
