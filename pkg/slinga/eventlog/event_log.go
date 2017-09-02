package eventlog

import (
	"github.com/Aptomi/aptomi/pkg/slinga/errors"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
)

type Fields map[string]interface{}

type AttachedObjects struct {
	objects []interface{}
}

type EventLog struct {
	*logrus.Logger
	attachedTo *AttachedObjects
	hook       *messageBuffer
}

// NewEventLog creates a new instance of event log
// Initially it just buffers all entries and doesn't write them
// It needs to buffer all entries, so that the context can be later attached to them
// before they get serialized and written to an external source
func NewEventLog() *EventLog {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	logger.Out = ioutil.Discard
	hook := &messageBuffer{}
	logger.Hooks.Add(hook)
	return &EventLog{
		Logger:     logger,
		attachedTo: &AttachedObjects{},
		hook:       hook,
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

	fields["attachedTo"] = eventLog.attachedTo
	return eventLog.Logger.WithFields(logrus.Fields(fields))
}

// AttachTo attaches all entries in this event log to a certain object
// E.g. dependency, user, service, context, serviceKey
func (eventLog *EventLog) AttachTo(object interface{}) {
	eventLog.attachedTo.objects = append(eventLog.attachedTo.objects, object)
}

// Append adds entries to the event logs
func (log *EventLog) Append(that *EventLog) {
	log.hook.entries = append(log.hook.entries, that.hook.entries...)
}

func (log *EventLog) LogError(err error) {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		log.WithFields(Fields(errWithDetails.Details())).Error(err.Error())
	} else {
		log.WithFields(Fields{}).Error(err.Error())
	}
}

func (log *EventLog) LogErrorAsWarning(err error) {
	errWithDetails, isErrorWithDetails := err.(*errors.ErrorWithDetails)
	if isErrorWithDetails {
		log.WithFields(Fields(errWithDetails.Details())).Warning(err.Error())
	} else {
		log.WithFields(Fields{}).Warning(err.Error())
	}
}

// Save takes all buffered entries and saves them
func (eventLog *EventLog) Save(hook logrus.Hook) {
	for _, e := range eventLog.hook.entries {
		hook.Fire(e)
	}
}
