package eventlog

import (
	"github.com/Sirupsen/logrus"
)

type Fields map[string]interface{}

type EventLog struct {
	*logrus.Logger
}

func NewEventLog() *EventLog {
	return &EventLog{
		Logger: logrus.New(),
	}
}

func (eventLog *EventLog) WithFields(fields Fields) *logrus.Entry {
	return eventLog.Logger.WithFields(logrus.Fields(fields))
}

func (log *EventLog) AttachToInstance(key string) {
	// TODO:
}


// TODO: attach to dependency
// TODO: attach to user?
// TODO: attach to componentId


// may be add contextual information automatically (service, component, etc) and attach it to all events
