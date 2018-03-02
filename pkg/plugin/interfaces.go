package plugin

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
)

// Registry is a registry of all Aptomi engine plugins
type Registry interface {
	ForCluster(cluster *lang.Cluster) (ClusterPlugin, error)
	ForCodeType(cluster *lang.Cluster, codeType string) (CodePlugin, error)
	PostProcess() []PostProcessPlugin
}

// RegistryFactory returns plugins registry on demand
type RegistryFactory func() Registry

// Base is a base interface for all engine plugins
type Base interface {
	Cleanup() error
}

// ClusterPlugin is a definition of cluster plugin which takes care of cluster operations such as validation
// in the cloud. It's created for specific cluster and enforcement cycle or API call.
type ClusterPlugin interface {
	Base

	Validate() error
}

// ClusterPluginConstructor represents constructor for the cluster plugin
type ClusterPluginConstructor func(cluster *lang.Cluster, cfg config.Plugins) (ClusterPlugin, error)

// CodePlugin is a definition of deployment plugin which takes care of creating, updating and destroying
// component instances in the cloud. It's created for specific cluster and enforcement cycle or API call.
type CodePlugin interface {
	Base

	Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error
	Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error)
	Status(deployName string, params util.NestedParameterMap, eventLog *event.Log) (DeploymentStatus, error)
}

// CodePluginConstructor represents constructor the the code plugin
type CodePluginConstructor func(cluster ClusterPlugin, cfg config.Plugins) (CodePlugin, error)

// PostProcessPlugin is a definition of post-processing plugin which gets called once by an action from the engine
// applier, after engine is done processing all component instances.
type PostProcessPlugin interface {
	Base

	Process(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, externalData *external.Data, eventLog *event.Log) error
}
