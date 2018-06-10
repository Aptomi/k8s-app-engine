package fake

import (
	"github.com/Aptomi/aptomi/pkg/plugin"
	"time"
)

type noOpPlugin struct {
	sleepTime time.Duration
}

var _ plugin.ClusterPlugin = &noOpPlugin{}
var _ plugin.CodePlugin = &noOpPlugin{}

// NewNoOpClusterPlugin returns fake cluster plugin which does nothing, except sleeping a given time amount on every action
func NewNoOpClusterPlugin(sleepTime time.Duration) plugin.ClusterPlugin {
	return &noOpPlugin{
		sleepTime: sleepTime,
	}
}

// NewNoOpCodePlugin returns fake code plugin which does nothing, except sleeping a given time amount on every action
func NewNoOpCodePlugin(sleepTime time.Duration) plugin.CodePlugin {
	return &noOpPlugin{
		sleepTime: sleepTime,
	}
}

func (plugin *noOpPlugin) Validate() error {
	return nil
}

func (plugin *noOpPlugin) Cleanup() error {
	return nil
}

func (plugin *noOpPlugin) Create(invocation *plugin.CodePluginInvocationParams) error {
	time.Sleep(plugin.sleepTime)
	return nil
}

func (plugin *noOpPlugin) Update(invocation *plugin.CodePluginInvocationParams) error {
	time.Sleep(plugin.sleepTime)
	return nil
}

func (plugin *noOpPlugin) Destroy(invocation *plugin.CodePluginInvocationParams) error {
	time.Sleep(plugin.sleepTime)
	return nil
}

func (plugin *noOpPlugin) Endpoints(invocation *plugin.CodePluginInvocationParams) (map[string]string, error) {
	time.Sleep(plugin.sleepTime)
	return map[string]string{
		"http": "endpoint_fake",
	}, nil
}

func (plugin *noOpPlugin) Resources(invocation *plugin.CodePluginInvocationParams) (plugin.Resources, error) {
	return nil, nil
}

func (plugin *noOpPlugin) Status(invocation *plugin.CodePluginInvocationParams) (bool, error) {
	return true, nil
}
