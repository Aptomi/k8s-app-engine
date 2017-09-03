package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
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

type EnginePluginImpl struct {
	failComponents []string
	eventLog       *eventlog.EventLog
}

func NewEnginePluginImpl(failComponents []string) *EnginePluginImpl {
	return &EnginePluginImpl{failComponents: failComponents}
}

func (p *EnginePluginImpl) OnApplyComponentInstanceCreate(key string) error {
	p.eventLog.Infof("[+] %s", key)
	for _, s := range p.failComponents {
		if strings.Contains(key, s) {
			return fmt.Errorf("Apply failed for component: " + key)
		}
	}
	return nil
}

func (p *EnginePluginImpl) OnApplyComponentInstanceUpdate(key string) error {
	p.eventLog.Infof("[*] %s", key)
	for _, s := range p.failComponents {
		if strings.Contains(key, s) {
			return fmt.Errorf("Update failed for component: " + key)
		}
	}
	return nil
}

func (p *EnginePluginImpl) OnApplyComponentInstanceDelete(key string) error {
	p.eventLog.Infof("[-] %s", key)
	for _, s := range p.failComponents {
		if strings.Contains(key, s) {
			return fmt.Errorf("Delete failed for component: " + key)
		}
	}
	return nil
}
