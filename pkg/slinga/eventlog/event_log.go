package eventlog

import (
	"github.com/Sirupsen/logrus"
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
)

type Fields map[string]interface{}

type EventLog struct {
	*logrus.Logger
	attachedToObjects []interface{}
}

// NewEventLog creates a new instance of event log
func NewEventLog() *EventLog {
	return &EventLog{
		Logger: logrus.New(),
	}
}

// WithFields creates a new log entry with a given set of fields
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

// AttachTo attaches all entries in this event log to a certain object
// E.g. dependency, user, service, context, serviceKey
func (eventLog *EventLog) AttachTo(object interface{}) {
	eventLog.attachedToObjects = append(eventLog.attachedToObjects, object)
}
