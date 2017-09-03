package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
)

type DeployPlugin interface {
	Plugin

	GetSupportedCodeTypes() []string
	Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error
	Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error
	Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error
	Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) (map[string]string, error)
}
