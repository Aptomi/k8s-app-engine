package eventlog

import (
	"github.com/Sirupsen/logrus"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
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
	// see if there is any details which have to added to the log from the error
	for _, value := range fields {
		if errWithDetails, ok := value.(*errors.ErrorWithDetails); ok {
			// put details from the error into the same log record
			for dKey, dValue := range errWithDetails.Details() {
				fields[dKey] = dValue
			}
		}
	}

	return eventLog.Logger.WithFields(logrus.Fields(fields))
}

func (log *EventLog) AttachToInstance(key string) {
	// TODO:
}

// TODO: attach to dependency
// TODO: attach to user?
// TODO: attach to component

// may be add contextual information automatically (service, component, etc) and attach it to all events
