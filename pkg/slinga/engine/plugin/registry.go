package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/istio"
)

func AllPlugins() []EnginePlugin {
	return []EnginePlugin{
		&istio.IstioRuleEnforcer{},
	}
}
