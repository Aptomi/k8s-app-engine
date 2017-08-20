package visibility

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"html"
)

type serviceInstanceNode struct {
	key             string
	service         *Service
	context         string
	contextWithKeys string
	instance        *engine.ComponentInstance
	primary         bool
}

func newServiceInstanceNode(key string, service *Service, context string, contextWithKeys string, instance *engine.ComponentInstance, primary bool) graphNode {
	return serviceInstanceNode{
		key:             key,
		service:         service,
		context:         context,
		contextWithKeys: contextWithKeys,
		instance:        instance,
		primary:         primary,
	}
}

func (n serviceInstanceNode) getIDPrefix() string {
	return "svcinst-"
}

func (n serviceInstanceNode) getGroup() string {
	if n.primary {
		return "serviceInstancePrimary"
	}
	return "serviceInstance"
}

func (n serviceInstanceNode) getID() string {
	return fmt.Sprintf("%s%s", n.getIDPrefix(), getWebIDByComponentKey(n.key))
}

func (n serviceInstanceNode) isItMyID(id string) string {
	return getWebComponentKeyByID(cutPrefixOrEmpty(id, n.getIDPrefix()))
}

func (n serviceInstanceNode) getLabel() string {
	// for not resolved instances
	if !n.instance.Resolved {
		if n.service == nil {
			return fmt.Sprintf(
				`<b>%s</b>
					ERROR`,
				html.EscapeString(n.instance.Key.ServiceName),
			)
		}

		return fmt.Sprintf(
			`<b>%s</b>
				ERROR`,
			html.EscapeString(n.service.GetName()),
		)
	}

	// for successfully resolved instances (primary & not primary)
	if n.primary {
		return fmt.Sprintf(
			`<b>%s</b>
				components: <i>%d</i>
				cluster: <i>%s</i>
				running: <i>%s</i>`,
			html.EscapeString(n.service.GetName()),
			len(n.service.Components), // TODO: fix
			html.EscapeString(n.instance.CalculatedLabels.Labels["cluster"]),
			html.EscapeString(NewTimeDiff(n.instance.GetRunningTime()).Humanize()),
		)
	}
	return fmt.Sprintf(
		`<b>%s</b>
			cluster: <i>%s</i>
			running: <i>%s</i>`,
		html.EscapeString(n.service.GetName()),
		html.EscapeString(n.instance.CalculatedLabels.Labels["cluster"]),
		html.EscapeString(NewTimeDiff(n.instance.GetRunningTime()).Humanize()),
	)
}

func (n serviceInstanceNode) getEdgeLabel(dst graphNode) string {
	return ""
}

func (n serviceInstanceNode) getDetails(id string, state engine.ServiceUsageState) interface{} {
	result := state.ResolvedData.ComponentInstanceMap[id]
	if result == nil {
		result = state.UnresolvedData.ComponentInstanceMap[id]
	}
	return result
}
