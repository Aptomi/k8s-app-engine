package event

import (
	"github.com/Sirupsen/logrus"
)

// HookMemory implements event log hook, which buffers all event log entries in memory
type HookMemory struct {
	entries []*logrus.Entry
}

// Levels defines on which log levels this hook should be fired
func (buf *HookMemory) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (buf *HookMemory) Fire(e *logrus.Entry) error {
	buf.entries = append(buf.entries, e)
	return nil
}
