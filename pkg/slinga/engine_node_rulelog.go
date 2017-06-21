package slinga

import "fmt"

// RuleLogType is a type of the log entry (e.g. rule evaluation, or just informational record)
type RuleLogType string

// RuleLogTypeDebug is for debug (something low-level happened)
const RuleLogTypeDebug RuleLogType = "Debug"

// RuleLogTypeInfo is for info records (something happened)
const RuleLogTypeInfo RuleLogType = "Info"

// RuleLogTypeTest is for test records (something is being tested)
const RuleLogTypeTest RuleLogType = "Test"

// RuleLogScope is a scope of the rule entry
type RuleLogScope string

// RuleLogScopeLocal is for local records (policy evaluation)
const RuleLogScopeLocal RuleLogScope = "Local"

// RuleLogScopeGlobal is for global records (global rules)
const RuleLogScopeGlobal RuleLogScope = "Global"

// RuleLogEntry is an entry that corresponds to rule
type RuleLogEntry struct {
	Type      RuleLogType
	Scope     RuleLogScope
	Name      string
	Message   string
	Condition string
	Result    bool
}

// NewRuleLogEntry creates new RuleLogEntry
func NewRuleLogEntry(ruleLogType RuleLogType, ruleLogScope RuleLogScope, name string, message string, condition string, result bool) *RuleLogEntry {
	return &RuleLogEntry{
		Type:      ruleLogType,
		Scope:     ruleLogScope,
		Name:      name,
		Message:   message,
		Condition: condition,
		Result:    result,
	}
}

// RuleLogWriter is something that is capable of organizing and storing rule logs in the right entities
type RuleLogWriter struct {
	// global structure, inside which logs will be saved
	resolvedUsage *ResolvedServiceUsageData

	// instance key where rule logs should be attached to (initially will be empty, but will be set as evaluation goes on)
	key string

	// dependency where rule logs should be attached to
	dependency *Dependency

	// queue of pending entries (will be filled while key is empty and until key is set)
	queue []*RuleLogEntry
}

// NewRuleLogWriter creates new RuleLogWriter for writing rule logs
func NewRuleLogWriter(resolvedUsage *ResolvedServiceUsageData, dependency *Dependency) *RuleLogWriter {
	return &RuleLogWriter{
		resolvedUsage: resolvedUsage,
		dependency:    dependency,
	}
}

func (writer *RuleLogWriter) setInstanceKey(key string) {
	if len(key) <= 0 {
		debug.Fatal("Empty instance key")
	}
	writer.key = key
	writer.flushQueue()
}

func (writer *RuleLogWriter) flushQueue() {
	// store all items from queue
	for _, entry := range writer.queue {
		writer.resolvedUsage.storeRuleLogEntry(writer.key, writer.dependency, entry)
	}
}

// Adds an entry into rule log
func (writer *RuleLogWriter) addRuleLogEntry(entry *RuleLogEntry) {
	if len(writer.key) <= 0 {
		// no key is set -> put into queue
		writer.queue = append(writer.queue, entry)
	} else {
		// store item directly
		writer.resolvedUsage.storeRuleLogEntry(writer.key, writer.dependency, entry)
	}
}

func entryResolvingDependencyStart(serviceName string, user *User) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Resolve (Dependency)",
		fmt.Sprintf("Resolving dependency for service '%s', user '%s' (ID = %s)", serviceName, user.Name, user.ID),
		"N/A",
		true,
	)
}

func entryResolvingDependencyEnd(service *Service, user *User) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Resolved (Dependency)",
		fmt.Sprintf("Successfully resolved dependency for service '%s', user '%s' (ID = %s)", service.Name, user.Name, user.ID),
		"N/A",
		true,
	)
}

func entryLabels(labels LabelSet) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeDebug,
		RuleLogScopeLocal,
		"Show (Labels)",
		fmt.Sprintf("Labels: '%s'", labels),
		"N/A",
		true,
	)
}

func entryContextsFound(service *Service, result bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeLocal,
		"Exist (Contexts)",
		fmt.Sprintf("Checking if contexts are present for service '%s'", service.Name),
		"has(contexts)",
		result,
	)
}

func entryContextCriteriaTesting(context *Context, matched bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeLocal,
		"Matches (Context)",
		fmt.Sprintf("Testing context (criteria): '%s'", context.Name),
		fmt.Sprintf("%+v", context.Criteria),
		matched,
	)
}

func entryContextMatched(service *Service, contextMatched *Context) *RuleLogEntry {
	var message string
	if contextMatched != nil {
		message = fmt.Sprintf("Context matched for service '%s': %s", service.Name, contextMatched.Name)
	} else {
		message = fmt.Sprintf("Unable to find matching context for service '%s'", service.Name)
	}

	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Matched (Context)",
		message,
		"N/A",
		contextMatched != nil,
	)
}

func entryAllocationsFound(service *Service, context *Context, result bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeLocal,
		"Exist (Allocations)",
		fmt.Sprintf("Checking if allocations are present for service '%s', context '%s'", service.Name, context.Name),
		"has(allocations)",
		result,
	)
}

func entryAllocationCriteriaTesting(allocation *Allocation, matched bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeLocal,
		"Matches (Allocation)",
		fmt.Sprintf("Testing allocation (criteria): '%s'", allocation.Name),
		fmt.Sprintf("%+v", allocation.Criteria),
		matched,
	)
}

func entryAllocationGlobalRuleTesting(allocation *Allocation, rule *Rule, matched bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeGlobal,
		"Global Rule (Allocation)",
		fmt.Sprintf("Testing if global rule '%s' applies to allocation '%s'", rule.Name, allocation.Name),
		fmt.Sprintf("%+v", rule.FilterServices),
		matched,
	)
}

func entryAllocationGlobalRulesNoViolations(allocation *Allocation, matched bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeGlobal,
		"No Global Rule Violations (Allocation)",
		fmt.Sprintf("Checking global rule violations for allocation: '%s'", allocation.Name),
		"!has(global_rule_violations)",
		matched,
	)
}

func entryAllocationMatched(service *Service, context *Context, allocationMatched *Allocation) *RuleLogEntry {
	var message string
	if allocationMatched != nil {
		message = fmt.Sprintf("Allocation matched for service '%s', context '%s': %s", service.Name, context.Name, allocationMatched.NameResolved)
	} else {
		message = fmt.Sprintf("Unable to find matching allocation for service '%s', context '%s'", service.Name, context.Name)
	}

	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Matched (Allocation)",
		message,
		"N/A",
		allocationMatched != nil,
	)
}
