package event

import (
	"github.com/Aptomi/aptomi/pkg/errors"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
)

// Fields is a set of named fields. Fields are attached to every log record
type Fields map[string]interface{}

// attachedObjects is a list of core aptomi lang objects attached to a set of log records (e.g. dependency, user, contract, context, key)
type attachedObjects struct {
	objects []interface{}
}

// Log is an buffered event log.
// It stores all log entries in memory first, then allows them to be processed and stored
type Log struct {
	logger     *logrus.Logger
	attachedTo *attachedObjects
	hookMemory *HookMemory
	scope      string
	log        bool
}

// NewLog creates a new instance of event log.
// Initially it just buffers all entries and doesn't write them.
// It needs to buffer all entries, so that the context can be later attached to them
// before they get serialized and written to an external source
func NewLog(scope string, log bool) *Log {
	logger := &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	hookMemory := &HookMemory{}
	logger.Hooks.Add(hookMemory)

	if log {
		logger.Hooks.Add(&HookLogger{})
	}

	return &Log{
		logger:     logger,
		attachedTo: &attachedObjects{},
		hookMemory: hookMemory,
		scope:      scope,
		log:        log,
	}
}

// Replaces base objects with their string key value
func fieldValue(data interface{}) interface{} {
	if baseObject, ok := data.(runtime.Storable); ok {
		return runtime.KeyForStorable(baseObject)
	}
	return data
}

// GetScope returns scope for event log
func (eventLog *Log) GetScope() string {
	return eventLog.scope
}

// IsLog returns true if current event log writes logs on the go
func (eventLog *Log) IsLog() bool {
	return eventLog.log
}

// WithFields creates a new log entry with a given set of fields
func (eventLog *Log) WithFields(fields Fields) *logrus.Entry {
	// see if there is any details which have to added to the log from the error
	for key, value := range fields {
		if errWithDetails, ok := value.(*errors.ErrorWithDetails); ok {
			// put details from the error into the same log record
			for dKey, dValue := range errWithDetails.Details() {
				fields[dKey] = fieldValue(dValue)
			}
			fields[key] = errWithDetails.Error()
		} else {
			fields[key] = fieldValue(value)
		}
	}

	if len(eventLog.scope) > 0 {
		fields["scope"] = eventLog.scope
	}

	return eventLog.logger.WithFields(logrus.Fields(fields))
}

// AttachTo attaches all entries in this event log to a certain object
// E.g. dependency, user, contract, context, serviceKey
func (eventLog *Log) AttachTo(object interface{}) {
	eventLog.attachedTo.objects = append(eventLog.attachedTo.objects, object)
}

// Append adds entries to the event logs
func (eventLog *Log) Append(that *Log) {
	if eventLog.IsLog() && !that.IsLog() {
		logger := &HookLogger{}
		for _, entry := range that.hookMemory.entries {
			err := logger.Fire(entry)
			if err != nil {
				logrus.Panicf("error while firing events during appending event logs: %s", err)
			}
		}
	}

	eventLog.hookMemory.entries = append(eventLog.hookMemory.entries, that.hookMemory.entries...)
}

// LogError logs an error. Errors with details are processed specially, their details get unfolded as record fields
func (eventLog *Log) LogError(err error) {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		eventLog.WithFields(Fields(errWithDetails.Details())).Error(err.Error())
	} else {
		eventLog.WithFields(Fields{}).Error(err.Error())
	}
}

// LogWarning logs a warning. Errors with details are processed specially, their details get unfolded as record fields
func (eventLog *Log) LogWarning(err error) {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		eventLog.WithFields(Fields(errWithDetails.Details())).Warning(err.Error())
	} else {
		eventLog.WithFields(Fields{}).Warning(err.Error())
	}
}

// Save takes all buffered event log entries and saves them
func (eventLog *Log) Save(hook logrus.Hook) {
	for _, e := range eventLog.hookMemory.entries {
		e.Data["attachedTo"] = eventLog.attachedTo
		err := hook.Fire(e)
		if err != nil {
			panic(err) // is it ok to panic here?
		}
	}
}
