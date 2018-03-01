package event

import (
	"github.com/Sirupsen/logrus"
	"time"
)

type APIEvent struct {
	Time     time.Time
	LogLevel string `yaml:"level"`
	Message  string
}

// AsAPIEvents takes all buffered event log entries and saves them as APIEvents
func (eventLog *Log) AsAPIEvents() []*APIEvent {
	saver := &HookApiEvents{}
	eventLog.Save(saver)

	return saver.events
}

// HookApiEvents saves all events as APIEvents that holds only time, level and message
type HookApiEvents struct {
	events []*APIEvent
}

// Levels defines on which log levels this hook should be fired
func (hook *HookApiEvents) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (hook *HookApiEvents) Fire(e *logrus.Entry) error {
	apiEvent := &APIEvent{e.Time, e.Level.String(), e.Message}
	hook.events = append(hook.events, apiEvent)
	return nil
}
