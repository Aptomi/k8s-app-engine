package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
)

// RuleLogType is a type of the log entry (e.g. rule evaluation, or just informational record)
type RuleLogType string

// RuleLogScope is a scope of the rule entry
type RuleLogScope string

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
		panic("Trying to attach logs to empty instance key")
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
