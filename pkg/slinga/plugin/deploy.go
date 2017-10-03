package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
)

type DeployPlugin interface {
	Plugin

	GetSupportedCodeTypes() []string
	Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error)
}
