package event

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

// HookBoltDB allows event log entries to be persisted in BoltDB
type HookBoltDB struct {
}

// Levels says that this hook should be fired on messages of all log levels
func (buf *HookBoltDB) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (buf *HookBoltDB) Fire(e *logrus.Entry) error {
	// figure out to which objects this entry should be attached to
	attachedToObjects := e.Data["attachedTo"].(*AttachedObjects).objects
	delete(e.Data, "attachedTo")

	// TODO: store this entry into bolt
	// attachedToObjects is a slice with attached objects (i.e. dependency, user, service, context, serviceKey)
	fmt.Printf("[%s] %s %p %p\n", e.Level, e.Message, e.Data, attachedToObjects)
	return nil
}
