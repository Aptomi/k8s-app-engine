package fake

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/util"
	"strings"
)

type failCodePlugin struct {
	// failComponents is a list of substrings to search in component names. When found, the corresponding component will be failed
	failComponents []string

	// failAsPanic, if set to true, will panic on matching components. Otherwise it will return an error
	failAsPanic bool
}

var _ plugin.CodePlugin = &failCodePlugin{}

// NewFailCodePlugin returns fake code plugin that does nothing, except fails component actions if their name contains
// one of the given strings
func NewFailCodePlugin(failComponents []string, failAsPanic bool) plugin.CodePlugin {
	return &failCodePlugin{
		failComponents: failComponents,
		failAsPanic:    failAsPanic,
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
	for _, s := range plugin.failComponents {
		if strings.Contains(deployName, s) {
			return plugin.fail("create", deployName)
		}
	}
	return nil
}

func (plugin *failCodePlugin) Update(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[*] %s", deployName)
	for _, s := range plugin.failComponents {
		if strings.Contains(deployName, s) {
			return plugin.fail("update", deployName)
		}
	}
	return nil
}

func (plugin *failCodePlugin) Destroy(deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[-] %s", deployName)
	for _, s := range plugin.failComponents {
		if strings.Contains(deployName, s) {
			return plugin.fail("delete", deployName)
		}
	}
	return nil
}

func (plugin *failCodePlugin) Endpoints(deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	return make(map[string]string), nil
}
