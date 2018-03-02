package k8s

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/client-go/pkg/api/v1"
	"strings"
)

func (plugin *Plugin) EndpointsForManifests(deployName, targetManifest string, eventLog *event.Log) (map[string]string, error) {
	kubeClient, err := plugin.NewClient()
	if err != nil {
		return nil, err
	}

	helmKube := plugin.NewHelmKube(deployName, eventLog)

	infos, err := helmKube.BuildUnstructured(plugin.Namespace, strings.NewReader(targetManifest))
	if err != nil {
		return nil, err
	}

	endpoints := make(map[string]string)

	for _, info := range infos {
		if info.Mapping.GroupVersionKind.Kind == "Service" {
			service, getErr := kubeClient.CoreV1().Services(plugin.Namespace).Get(info.Name, meta.GetOptions{})
			if getErr != nil {
				return nil, getErr
			}

			plugin.addEndpointsFromService(service, endpoints)
		}
	}

	return endpoints, nil
}

// addEndpointsFromService searches for the available endpoints in specified service and writes them into provided map
func (plugin *Plugin) addEndpointsFromService(service *api.Service, endpoints map[string]string) {
	// todo(slukjanov): support not only node ports
	if service.Spec.Type == "NodePort" {
		for _, port := range service.Spec.Ports {
			sURL := fmt.Sprintf("%s:%d", plugin.ExternalAddress, port.NodePort)

			// todo(slukjanov): could we somehow detect real schema? I think no :(
			if util.StringContainsAny(port.Name, "https") {
				sURL = "https://" + sURL
			} else if util.StringContainsAny(port.Name, "ui", "rest", "http", "grafana") {
				sURL = "http://" + sURL
			}

			name := port.Name
			if len(name) == 0 {
				name = port.TargetPort.String()
			}

			endpoints[name] = sURL
		}
	}
}
