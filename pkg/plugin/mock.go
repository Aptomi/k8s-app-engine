package plugin

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"strings"
	"time"
)

// MockRegistry is a mock plugin registry, with a single deployment plugin and a single post-processing plugin.
// It's useful in unit tests and running policy apply in noop mode (e.g. for testing UI without deploying changes)
type MockRegistry struct {
	DeployPlugin      DeployPlugin
	PostProcessPlugin PostProcessPlugin
}

// GetDeployPlugin always returns the same deployment plugin
func (reg *MockRegistry) GetDeployPlugin(codeType string) (DeployPlugin, error) {
	return reg.DeployPlugin, nil
}

// GetPostProcessingPlugins always returns the same post-processing plugin
func (reg *MockRegistry) GetPostProcessingPlugins() []PostProcessPlugin {
	return []PostProcessPlugin{reg.PostProcessPlugin}
}

// MockDeployPlugin is a mock plugin which does nothing, except sleeping a given time amount on every action
type MockDeployPlugin struct {
	SleepTime time.Duration
}

// Cleanup does nothing
func (p *MockDeployPlugin) Cleanup() error {
	return nil
}

// GetSupportedCodeTypes does nothing
func (p *MockDeployPlugin) GetSupportedCodeTypes() []string {
	return []string{}
}

// Create does nothing but sleeps
func (p *MockDeployPlugin) Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	time.Sleep(p.SleepTime)
	return nil
}

// Update does nothing but sleeps
func (p *MockDeployPlugin) Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	time.Sleep(p.SleepTime)
	return nil
}

// Destroy does nothing but sleeps
func (p *MockDeployPlugin) Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	time.Sleep(p.SleepTime)
	return nil
}

// Endpoints sleeps and then always returns an empty set of endpoints
func (p *MockDeployPlugin) Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	time.Sleep(p.SleepTime)
	return make(map[string]string), nil
}

// MockDeployPluginFailComponents is a mock plugin which does nothing, except fails component actions if their name contains
// one of the given strings
type MockDeployPluginFailComponents struct {
	// FailComponents is a list of substrings to search in component names. When found, the corresponding component will be failed
	FailComponents []string

	// FailAsPanic, if set to true, will panic on matching components. Otherwise it will return an error
	FailAsPanic bool
}

// Cleanup does nothing
func (p *MockDeployPluginFailComponents) Cleanup() error {
	return nil
}

// GetSupportedCodeTypes does nothing
func (p *MockDeployPluginFailComponents) GetSupportedCodeTypes() []string {
	return []string{}
}

// Create does nothing, except failing components if their name contains of the strings from FailComponents
func (p *MockDeployPluginFailComponents) Create(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[+] %s", deployName)
	for _, s := range p.FailComponents {
		if strings.Contains(deployName, s) {
			return p.fail("create", deployName)
		}
	}
	return nil
}

// Update does nothing, except failing components if their name contains of the strings from FailComponents
func (p *MockDeployPluginFailComponents) Update(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[*] %s", deployName)
	for _, s := range p.FailComponents {
		if strings.Contains(deployName, s) {
			return p.fail("update", deployName)
		}
	}
	return nil
}

// Destroy does nothing, except failing components if their name contains of the strings from FailComponents
func (p *MockDeployPluginFailComponents) Destroy(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) error {
	eventLog.WithFields(event.Fields{}).Infof("[-] %s", deployName)
	for _, s := range p.FailComponents {
		if strings.Contains(deployName, s) {
			return p.fail("delete", deployName)
		}
	}
	return nil
}

func (p *MockDeployPluginFailComponents) fail(action string, deployName string) error {
	msg := fmt.Sprintf("%s failed by plugin mock for component '%s' (panic = %t)", action, deployName, p.FailAsPanic)
	if p.FailAsPanic {
		panic(msg)
	}
	return fmt.Errorf(msg)
}

// Endpoints always returns an empty set of endpoints
func (p *MockDeployPluginFailComponents) Endpoints(cluster *lang.Cluster, deployName string, params util.NestedParameterMap, eventLog *event.Log) (map[string]string, error) {
	return make(map[string]string), nil
}

// MockPostProcessPlugin is a mock post-processing plugin which does nothing
type MockPostProcessPlugin struct {
}

// Process does nothing
func (p *MockPostProcessPlugin) Process(desiredPolicy *lang.Policy, desiredState *resolve.PolicyResolution, externalData *external.Data, eventLog *event.Log) error {
	return nil
}

// Cleanup does nothing
func (p *MockPostProcessPlugin) Cleanup() error {
	return nil
}
