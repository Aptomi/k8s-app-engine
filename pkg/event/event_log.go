package event

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

// Fields is a set of named fields. Fields are attached to every log record
type Fields map[string]interface{}

// Log is an buffered event log.
// It stores all log entries in memory first, then allows them to be processed and stored
type Log struct {
	logger      *logrus.Logger
	hookMemory  *HookMemory
	level       logrus.Level
	scope       string
	fixedFields map[string]string
}

// NewLog creates a new instance of event log.
// Initially it just buffers all entries and doesn't write them.
// It needs to buffer all entries, so that the context can be later attached to them
// before they get serialized and written to an external source
func NewLog(level logrus.Level, scope string) *Log {
	logger := &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	hookMemory := &HookMemory{}
	logger.Hooks.Add(hookMemory)

	return &Log{
		logger:     logger,
		hookMemory: hookMemory,
		level:      level,
		scope:      scope,
		fixedFields: map[string]string{
			"scope": scope,
		},
	}
}

// AddHook puts an additional hook to an existing event log
func (eventLog *Log) AddHook(hook logrus.Hook) *Log {
	eventLog.logger.Hooks.Add(hook)
	return eventLog
}

// AddConsoleHook puts an additional hook to an existing event log, to mirror logs to the console
func (eventLog *Log) AddConsoleHook(level logrus.Level) *Log {
	return eventLog.AddHook(NewHookConsole(level))
}

// GetLevel returns log level for the event log
func (eventLog *Log) GetLevel() logrus.Level {
	return eventLog.level
}

// GetScope returns scope for event log
func (eventLog *Log) GetScope() string {
	return eventLog.scope
}

// NewEntry creates a new log entry
func (eventLog *Log) NewEntry() *logrus.Entry {
	logRusFields := logrus.Fields{}

	// add fixed fields
	for key, value := range eventLog.fixedFields {
		logRusFields[key] = value
	}

	// pass to logrus
	return eventLog.logger.WithFields(logRusFields)
}

// Append adds entries to the event logs
func (eventLog *Log) Append(that *Log) {
	for _, thatEntry := range that.hookMemory.entries {
		entry := &logrus.Entry{
			Logger:  eventLog.logger,
			Data:    thatEntry.Data,
			Time:    thatEntry.Time,
			Level:   thatEntry.Level,
			Message: thatEntry.Message,
		}
		switch entry.Level {
		case logrus.PanicLevel:
			entry.Panic(entry.Message)
		case logrus.FatalLevel:
			entry.Fatal(entry.Message)
		case logrus.ErrorLevel:
			entry.Error(entry.Message)
		case logrus.WarnLevel:
			entry.Warn(entry.Message)
		case logrus.InfoLevel:
			entry.Info(entry.Message)
		case logrus.DebugLevel:
			entry.Debug(entry.Message)
		}
	}
}

// AddFixedField adds field with name=values to add following entries in the log
func (eventLog *Log) AddFixedField(name string, value string) {
	eventLog.fixedFields[name] = value
}

// Save takes all buffered event log entries and saves them
func (eventLog *Log) Save(hook logrus.Hook) {
	for _, e := range eventLog.hookMemory.entries {
		err := hook.Fire(e)
		if err != nil {
			panic(err)
		}
	}
}
