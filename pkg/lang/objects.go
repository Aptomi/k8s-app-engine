package lang

import "github.com/Aptomi/aptomi/pkg/runtime"

var (
	// PolicyTypes is the list of informational data for all policy objects
	PolicyTypes = []*runtime.TypeInfo{
		TypeBundle,
		TypeService,
		TypeClaim,
		TypeCluster,
		TypeRule,
		TypeACLRule,
	}

	policyObjectsMap = make(map[runtime.Kind]bool)
)

func init() {
	for _, obj := range PolicyTypes {
		policyObjectsMap[obj.Kind] = true
	}
}

// IsPolicyObject returns true if provided object is part of the policy objects list
func IsPolicyObject(obj runtime.Object) bool {
	return policyObjectsMap[obj.GetKind()]
}
