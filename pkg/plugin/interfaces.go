package plugin

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
)

// Registry is a registry of all Aptomi engine plugins
type Registry interface {
	ForCluster(cluster *lang.Cluster) (ClusterPlugin, error)
	ForCodeType(cluster *lang.Cluster, codeType string) (CodePlugin, error)
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

	Create(*CodePluginInvocationParams) error
	Update(*CodePluginInvocationParams) error
	Destroy(*CodePluginInvocationParams) error
	Endpoints(*CodePluginInvocationParams) (map[string]string, error)
	Resources(*CodePluginInvocationParams) (Resources, error)
	Status(*CodePluginInvocationParams) (bool, error)
}

// ParamTargetSuffix it's a plugin-specific parameter, which is additionally specifies where the code should reside (in case of k8s and Helm, it's a string consisting of k8s namespace)
const ParamTargetSuffix = "target-suffix"

// CodePluginInvocationParams is a struct that will be passed into CodePlugin when invoking its methods
type CodePluginInvocationParams struct {
	DeployName   string
	Params       util.NestedParameterMap
	PluginParams map[string]string
	EventLog     *event.Log
}

// CodePluginConstructor represents constructor the the code plugin
type CodePluginConstructor func(cluster ClusterPlugin, cfg config.Plugins) (CodePlugin, error)
