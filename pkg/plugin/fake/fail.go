package fake

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/util"
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

func (plugin *failCodePlugin) Create(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[+] %s", deployName)
	return plugin.fail("create", deployName)
}

func (plugin *failCodePlugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[*] %s", deployName)
	return plugin.fail("update", deployName)
}

func (plugin *failCodePlugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[-] %s", deployName)
	return plugin.fail("delete", deployName)
}

func (plugin *failCodePlugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	return make(map[string]string), nil
}

func (plugin *failCodePlugin) Resources(deployName string, params util.NestedParameterMap, eventLog *event.Log) (plugin.Resources, error) {
	return nil, nil
}
