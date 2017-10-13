package acl

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"sync"
)

type Resolver struct {
	rules     []*Rule
	cache     *expression.Cache
	roleCache sync.Map
}

func NewResolver(rules []*Rule) *Resolver {
	return &Resolver{
		rules:     rules,
		cache:     expression.NewCache(),
		roleCache: sync.Map{},
	}
}

func (resolver *Resolver) GetUserRole(user *lang.User) (*Role, error) {
	resultCached, ok := resolver.roleCache.Load(user.ID)
	if ok {
		return resultCached.(*Role), nil
	}

	result := lang.NewRuleActionResult(lang.NewLabelSet(make(map[string]string)))
	params := expression.NewParams(user.Labels, nil)
	for _, rule := range resolver.rules {
		matched, err := rule.Matches(params, resolver.cache)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve role for user '%s': %s", user.ID, err)
		}
		if matched {
			rule.ApplyActions(result)
			if rule.Actions.Stop {
				break
			}
		}
	}

	roleID := result.Labels.Labels[LabelRole]
	role, ok := Roles[roleID]
	if !ok {
		return Nobody, nil
	}

	resolver.roleCache.Store(user.ID, role)
	return role, nil
}

func (resolver *Resolver) GetUserPrivileges(user *lang.User) (*Role, error) {
}
