package plugin

import (
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
)

// DeployPlugin is a definition of deployment plugin which takes care of creating, updating and destroying
// component instances in the cloud
type DeployPlugin interface {
	Plugin

	GetSupportedCodeTypes() []string
	Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error)
}
