package lang

import (
	"strings"
)

// ACLRuleActions is a set of actions that can be performed by a ACL rule, assigning permissions to access namespaces
type ACLRuleActions struct {
	// AddRole is a map with role ID as key, while value is a set of comma-separated namespaces to which this role applies
	AddRole map[string]string `yaml:"add-role,omitempty" validate:"omitempty,addRoleNS"`
}

// ApplyActions applies rule actions and updates result
func (rule *ACLRule) ApplyActions(roleMap map[string]map[string]bool) {
	for roleID, namespaceList := range rule.Actions.AddRole {
		role := ACLRolesMap[roleID]
		if role == nil {
			// skip non-existing roles
			continue
		}

		nsMap := roleMap[roleID]
		if nsMap == nil {
			nsMap = make(map[string]bool)
			roleMap[roleID] = nsMap
		}

		// mark all namespaces for the role
		namespaces := strings.Split(namespaceList, ",")
		for _, namespace := range namespaces {
			nsMap[strings.TrimSpace(namespace)] = true
		}

		// if role covers all namespaces, mark it as well
		if role.Privileges.AllNamespaces {
			nsMap[namespaceAll] = true
		}
	}
}
