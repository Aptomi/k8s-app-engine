package plugin

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/base"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/deployment"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/istio"
)

func AllPlugins() []EnginePlugin {
	return []EnginePlugin{
		&deployment.DeployerPlugin{&base.BasePlugin{}},
		&istio.RuleEnforcerPlugin{&base.BasePlugin{}},
	}
}
