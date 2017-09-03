package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func getPolicy() *language.PolicyNamespace {
	return language.LoadUnitTestsPolicy("../../testdata/unittests")
}

func getExternalData() *external.Data {
	return external.NewData(
		users.NewUserLoaderFromDir("../../testdata/unittests"),
		secrets.NewSecretLoaderFromDir("../../testdata/unittests"),
	)
}

func resolvePolicy(t *testing.T, policy *language.PolicyNamespace, externalData *external.Data) *resolve.PolicyResolution {
	resolver := resolve.NewPolicyResolver(policy, externalData)
	result, eventLog, err := resolver.ResolveAllDependencies()
	if !assert.Nil(t, err, "Policy should be resolved without errors") {
		hook := &eventlog.HookStdout{}
		eventLog.Save(hook)
		t.FailNow()
	}

	return result
}

func applyAndCheck(t *testing.T, apply *EngineApply, expectedResult int, errorCnt int, errorMsg string) *resolve.PolicyResolution {
	actualState, eventLog, err := apply.Apply()

	if !assert.Equal(t, expectedResult != ResError, err == nil, "Apply status (success vs. error)") {
		// print log into stdout and exit
		hook := &eventlog.HookStdout{}
		eventLog.Save(hook)
		t.FailNow()
	}

	if expectedResult == ResError {
		// check for error messages
		verifier := eventlog.NewUnitTestLogVerifier(errorMsg)
		eventLog.Save(verifier)
		if !assert.Equal(t, errorCnt, verifier.MatchedErrorsCount(), "Apply event log should have correct number of error messages containing words: "+errorMsg) {
			hook := &eventlog.HookStdout{}
			eventLog.Save(hook)
			t.FailNow()
		}
	}
	return actualState
}

type componentTimes struct {
	created time.Time
	updated time.Time
}

func getTimes(t *testing.T, key string, u2 *resolve.PolicyResolution) componentTimes {
	return componentTimes{
		created: getInstanceInternal(t, key, u2).CreatedOn,
		updated: getInstanceInternal(t, key, u2).UpdatedOn,
	}
}

func getInstanceInternal(t *testing.T, key string, resolution *resolve.PolicyResolution) *resolve.ComponentInstance {
	instance, ok := resolution.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance exists in resolution data: "+key) {
		t.FailNow()
	}
	return instance
}

func getInstanceKey(serviceName string, contextName string, allocationKeysResolved []string, componentName string, policy *language.PolicyNamespace) string {
	return resolve.NewComponentInstanceKey(serviceName, policy.Contexts[contextName], allocationKeysResolved, policy.Services[serviceName].GetComponentsMap()[componentName]).GetKey()
}

func NewTestPluginRegistry(failComponents ...string) plugin.Registry {
	return &testRegistry{&testPlugin{failComponents}}
}

type testRegistry struct {
	*testPlugin
}

func (reg *testRegistry) GetDeployPlugin(codeType string) (plugin.DeployPlugin, error) {
	return reg.testPlugin, nil
}

func (reg *testRegistry) GetClustersPostProcessingPlugins() []plugin.ClustersPostProcessPlugin {
	return []plugin.ClustersPostProcessPlugin{reg.testPlugin}
}

type testPlugin struct {
	failComponents []string
}

func (p *testPlugin) GetSupportedCodeTypes() []string {
	return []string{}

}
func (p *testPlugin) Create(cluster *language.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error {
	eventLog.Infof("[+] %s", deployName)
	for _, s := range p.failComponents {
		if strings.Contains(deployName, s) {
			return fmt.Errorf("Apply failed for component: %s", deployName)
		}
	}
	return nil
}
func (p *testPlugin) Update(cluster *language.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error {
	eventLog.Infof("[*] %s", deployName)
	for _, s := range p.failComponents {
		if strings.Contains(deployName, s) {
			return fmt.Errorf("Update failed for component: %s", deployName)
		}
	}
	return nil
}
func (p *testPlugin) Destroy(cluster *language.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) error {
	eventLog.Infof("[-] %s", deployName)
	for _, s := range p.failComponents {
		if strings.Contains(deployName, s) {
			return fmt.Errorf("Delete failed for component: %s", deployName)
		}
	}
	return nil
}
func (p *testPlugin) Endpoints(cluster *language.Cluster, deployName string, params util.NestedParameterMap, eventLog *eventlog.EventLog) (map[string]string, error) {
	return nil, nil
}

func (p *testPlugin) Process(desiredPolicy *language.PolicyNamespace, desiredState *resolve.PolicyResolution, externalData *external.Data, eventLog *eventlog.EventLog) error {
	return nil
}
