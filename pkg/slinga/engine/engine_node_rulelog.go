package engine

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
)

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
	Type          RuleLogType
	Scope         RuleLogScope
	Name          string
	Message       string
	Condition     string
	Result        bool
	TerminalError bool
}

// NewRuleLogEntry creates new RuleLogEntry
func NewRuleLogEntry(ruleLogType RuleLogType, ruleLogScope RuleLogScope, name string, message string, condition string, result bool, terminalError bool) *RuleLogEntry {
	return &RuleLogEntry{
		Type:          ruleLogType,
		Scope:         ruleLogScope,
		Name:          name,
		Message:       message,
		Condition:     condition,
		Result:        result,
		TerminalError: terminalError,
	}
}

// RuleLogWriter is something that is capable of organizing and storing rule logs in the right entities
type RuleLogWriter struct {
	// reference to the global structure, inside which logs will be saved
	data *ServiceUsageData

	// instance key where rule logs should be attached to (initially will be empty, but will be set as evaluation goes on)
	key string

	// dependency where rule logs should be attached to
	dependency *Dependency

	// queue of pending entries (will be filled while key is empty and until key is set)
	queue []*RuleLogEntry
}

// NewRuleLogWriter creates new RuleLogWriter for writing rule logs
func NewRuleLogWriter(data *ServiceUsageData, dependency *Dependency) *RuleLogWriter {
	return &RuleLogWriter{
		data:       data,
		dependency: dependency,
	}
}

func (writer *RuleLogWriter) attachToInstance(key string) {
	if len(key) <= 0 {
		Debug.Panic("Empty instance key")
	}
	writer.key = key
	writer.flushQueue()
}

func (writer *RuleLogWriter) flushQueue() {
	// store all items from queue
	for _, entry := range writer.queue {
		writer.data.storeRuleLogEntry(writer.key, writer.dependency, entry)
	}
}

// Adds an entry into rule log
func (writer *RuleLogWriter) addRuleLogEntry(entry *RuleLogEntry) {
	if entry == nil {
		return
	}
	if len(writer.key) <= 0 {
		// no key is set -> put into queue
		writer.queue = append(writer.queue, entry)
	} else {
		// store item directly
		writer.data.storeRuleLogEntry(writer.key, writer.dependency, entry)
	}
}

func entryResolvingDependencyStart(serviceName string, user *User, dependency *Dependency) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Resolve (Dependency)",
		fmt.Sprintf("Resolving '%s' -> '%s', depends on '%s'", user.Name, dependency.Service, serviceName),
		"N/A",
		true,
		false,
	)
}

func entryResolvingDependencyEnd(serviceName string, user *User, dependency *Dependency) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Resolved (Dependency)",
		fmt.Sprintf("Successfully resolved '%s' -> '%s', dependency on '%s'", user.Name, dependency.Service, serviceName),
		"N/A",
		true,
		false,
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
		false,
	)
}

func entryServiceMatched(serviceName string, found bool) *RuleLogEntry {
	if !found {
		return NewRuleLogEntry(
			RuleLogTypeInfo,
			RuleLogScopeLocal,
			"Found (Service)",
			fmt.Sprintf("Unable to find service '%s'", serviceName),
			"N/A",
			false,
			true,
		)
	}
	return nil
}

func entryContextsFound(result bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeLocal,
		"Exist (Contexts)",
		fmt.Sprintf("Checking if contexts are present in the namespace"),
		"has(contexts)",
		result,
		false,
	)
}

func entryContextCriteriaTesting(context *Context, matched bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeLocal,
		"Matches (Context)",
		fmt.Sprintf("Testing context (criteria): '%s'", context.GetName()),
		fmt.Sprintf("%+v", context.Criteria),
		matched,
		false,
	)
}

func entryContextMatched(service *Service, contextMatched *Context) *RuleLogEntry {
	var message string
	if contextMatched != nil {
		message = fmt.Sprintf("Context matched for service '%s': %s", service.GetName(), contextMatched.GetName())
	} else {
		message = fmt.Sprintf("Unable to find matching context for service '%s'", service.GetName())
	}

	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Matched (Context)",
		message,
		"N/A",
		contextMatched != nil,
		contextMatched == nil,
	)
}

func entryContextGlobalRuleTesting(context *Context, rule *Rule, matched bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeGlobal,
		"Global Rule (Context)",
		fmt.Sprintf("Testing if global rule '%s' applies to context '%s'", rule.GetName(), context.Metadata.Name),
		fmt.Sprintf("%+v", rule.FilterServices),
		matched,
		false,
	)
}

func entryContextGlobalRulesNoViolations(context *Context, noViolations bool) *RuleLogEntry {
	return NewRuleLogEntry(
		RuleLogTypeTest,
		RuleLogScopeGlobal,
		"No Global Rule Violations (Context)",
		fmt.Sprintf("Verify there are no global rule violations for context: '%s'", context.Metadata.Name),
		"!has(global_rule_violations)",
		noViolations,
		false,
	)
}

func entryAllocationMatched(service *Service, context *Context, allocationNameResolved string) *RuleLogEntry {
	var message string
	if len(allocationNameResolved) > 0 {
		message = fmt.Sprintf("Allocation matched for service '%s', context '%s': %s", service.GetName(), context.GetName(), allocationNameResolved)
	} else {
		message = fmt.Sprintf("Unable to find matching allocation for service '%s', context '%s'", service.GetName(), context.GetName())
	}

	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Matched (Allocation)",
		message,
		"N/A",
		len(allocationNameResolved) > 0,
		len(allocationNameResolved) == 0,
	)
}
