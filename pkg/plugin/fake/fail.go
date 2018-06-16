package fake

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/plugin"
)

// failCodePlugin is a plugin which fails all of its actions
type failCodePlugin struct {
	// failAsPanic, if set to true, will panic on matching components. Otherwise it will return an error
	failAsPanic bool
}

var _ plugin.CodePlugin = &failCodePlugin{}

// NewFailCodePlugin returns fake code plugin that does nothing, except fails component actions if their deploy name
// contains one of the given strings
func NewFailCodePlugin(failAsPanic bool) plugin.CodePlugin {
	return &failCodePlugin{
		failAsPanic: failAsPanic,
	}
}

func (plugin *failCodePlugin) Cleanup() error {
	return nil
}

func (plugin *failCodePlugin) fail(action string, deployName string) error {
	msg := fmt.Sprintf("%s failed by plugin mock for component '%s' (panic = %t)", action, deployName, plugin.failAsPanic)
	if plugin.failAsPanic {
		panic(msg)
	}

	return fmt.Errorf(msg)
}

func (plugin *failCodePlugin) Create(invocation *plugin.CodePluginInvocationParams) error {
	invocation.EventLog.NewEntry().Infof("[+] %s", invocation.DeployName)
	return plugin.fail("create", invocation.DeployName)
}

func (plugin *failCodePlugin) Update(invocation *plugin.CodePluginInvocationParams) error {
	invocation.EventLog.NewEntry().Infof("[*] %s", invocation.DeployName)
	return plugin.fail("update", invocation.DeployName)
}

func (plugin *failCodePlugin) Destroy(invocation *plugin.CodePluginInvocationParams) error {
	invocation.EventLog.NewEntry().Infof("[-] %s", invocation.DeployName)
	return plugin.fail("delete", invocation.DeployName)
}

func (plugin *failCodePlugin) Endpoints(invocation *plugin.CodePluginInvocationParams) (map[string]string, error) {
	return make(map[string]string), nil
}

func (plugin *failCodePlugin) Resources(invocation *plugin.CodePluginInvocationParams) (plugin.Resources, error) {
	return nil, nil
}

func (plugin *failCodePlugin) Status(invocation *plugin.CodePluginInvocationParams) (bool, error) {
	return false, nil
}
