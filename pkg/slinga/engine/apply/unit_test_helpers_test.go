package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func getPolicy() *language.PolicyNamespace {
	return language.LoadUnitTestsPolicy("../../testdata/unittests")
}

func getUserLoader() language.UserLoader {
	return language.NewUserLoaderFromDir("../../testdata/unittests")
}

func resolvePolicy(t *testing.T, policy *language.PolicyNamespace, userLoader language.UserLoader) *resolve.PolicyResolution {
	resolver := resolve.NewPolicyResolver(policy, userLoader)
	result, err := resolver.ResolveAllDependencies()
	if !assert.Nil(t, err, "Policy should be resolved without errors") {
		t.FailNow()
	}
	return result
}

func applyAndCheck(t *testing.T, apply *EngineApply, expectedResult int, errorCnt int, errorMsg string) {
	err := apply.Apply()
	if !assert.Equal(t, expectedResult != ResError, err == nil, "Apply status (success vs. error)") {
		// print log into stdout and exit
		hook := &eventlog.HookStdout{}
		apply.eventLog.Save(hook)
		t.FailNow()
	}

	if expectedResult == ResError {
		// check for error messages
		verifier := eventlog.NewUnitTestLogVerifier(errorMsg)
		apply.eventLog.Save(verifier)
		if !assert.Equal(t, errorCnt, verifier.MatchedErrorsCount(), "Apply event log should have correct number of error messages containing words: "+errorMsg) {
			hook := &eventlog.HookStdout{}
			apply.eventLog.Save(hook)
			t.FailNow()
		}
	}
}

type EnginePluginImpl struct {
	failComponents []string
	eventLog       *eventlog.EventLog
}

func NewEnginePluginImpl(failComponents []string) *EnginePluginImpl {
	return &EnginePluginImpl{failComponents: failComponents}
}

func (p *EnginePluginImpl) Init(desiredPolicy *language.PolicyNamespace, desiredState *resolve.PolicyResolution, actualPolicy *language.PolicyNamespace, actualState *resolve.PolicyResolution, userLoader language.UserLoader, eventLog *eventlog.EventLog) {
	p.eventLog = eventLog
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

func (p *EnginePluginImpl) GetCustomApplyProgressLength() int {
	return 0
}

func (p *EnginePluginImpl) OnApplyCustom(progress progress.ProgressIndicator) error {
	return nil
}
