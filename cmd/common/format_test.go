package common

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/global"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormat_Text(t *testing.T) {
	cfg := &config.Client{Output: Text}

	{
		// with policy changes
		result := makePolicyUpdateResult(true)
		data, err := Format(cfg, true, result)
		assert.Nil(t, err, "Format should work without error")
		assert.Equal(t, "Policy Changes\tInstance Changes                                      \nGen 41 -> 42  \t[*] cluster#ns#contract#context#keysresolved#component\n              \t[+] cluster#ns#contract#context#keysresolved#component\n              \t[-] cluster#ns#contract#context#keysresolved#component",
			string(data), "Format should return expected table")
		// fmt.Println(string(data))
	}
	{
		// without policy changes
		result := makePolicyUpdateResult(false)
		data, err := Format(cfg, true, result)
		assert.Nil(t, err, "Format should work without error")
		assert.Equal(t, "Policy Changes\tInstance Changes                                      \nGen 42 (none) \t[*] cluster#ns#contract#context#keysresolved#component\n              \t[+] cluster#ns#contract#context#keysresolved#component\n              \t[-] cluster#ns#contract#context#keysresolved#component",
			string(data), "Format should return expected table")
		// fmt.Println(string(data))
	}
	{
		// empty set of actions
		result := &api.PolicyUpdateResult{
			PolicyGeneration: 42,
			Actions:          []string{},
		}
		data, err := Format(cfg, true, result)
		assert.Nil(t, err, "Format should work without error")
		assert.Equal(t, "Policy Changes\tInstance Changes\nGen 42 (none) \t(none)          ",
			string(data), "Format should return expected table")
		// fmt.Println(string(data))
	}
}

func makePolicyUpdateResult(policyChanged bool) *api.PolicyUpdateResult {
	key := resolve.NewComponentInstanceKey(
		&lang.Cluster{Metadata: lang.Metadata{Name: "cluster"}},
		&lang.Contract{Metadata: lang.Metadata{Name: "contract", Namespace: "ns"}},
		&lang.Context{Name: "context"},
		[]string{"keysresolved"},
		&lang.Service{Metadata: lang.Metadata{Name: "service"}},
		&lang.ServiceComponent{Name: "component"},
	)
	result := &api.PolicyUpdateResult{
		PolicyGeneration: 42,
		PolicyChanged:    policyChanged,
		Actions: []string{
			component.NewCreateAction(key.GetKey()).GetName(),
			component.NewUpdateAction(key.GetKey()).GetName(),
			component.NewDeleteAction(key.GetKey()).GetName(),
			component.NewDetachDependencyAction(key.GetKey(), "depId").GetName(),
			component.NewAttachDependencyAction(key.GetKey(), "depId").GetName(),
			component.NewEndpointsAction(key.GetKey()).GetName(),
			global.NewPostProcessAction().GetName(),
		},
	}
	return result
}
