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
	cik *ComponentInstanceKey

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

func (writer *RuleLogWriter) attachToInstance(cik *ComponentInstanceKey) {
	if cik == nil {
		Debug.Panic("Empty instance key")
	}
	writer.cik = cik
	writer.flushQueue()
}

func (writer *RuleLogWriter) flushQueue() {
	// store all items from queue
	for _, entry := range writer.queue {
		writer.data.storeRuleLogEntry(writer.cik, writer.dependency, entry)
	}
}

// Adds an entry into rule log
func (writer *RuleLogWriter) addRuleLogEntry(entry *RuleLogEntry) {
	if entry == nil {
		return
	}
	if writer.cik == nil {
		// no key is set -> put into queue
		writer.queue = append(writer.queue, entry)
	} else {
		// store item directly
		writer.data.storeRuleLogEntry(writer.cik, writer.dependency, entry)
	}
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

func entryAllocationKeysResolved(service *Service, context *Context, allocationKeys []string) *RuleLogEntry {
	var message string
	if len(allocationKeys) > 0 {
		message = fmt.Sprintf("Allocation keys resolved for service '%s', context '%s': %v", service.Name, context.Name, allocationKeys)
	}

	return NewRuleLogEntry(
		RuleLogTypeInfo,
		RuleLogScopeLocal,
		"Keys Resolution (Allocation)",
		message,
		"N/A",
		true,
		false,
	)
}
