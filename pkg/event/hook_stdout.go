package event

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

// HookStdout allows event log entries to be printed to stdout
type HookStdout struct {
}

// Levels says that this hook should be fired on messages of all log levels
func (buf *HookStdout) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (buf *HookStdout) Fire(e *logrus.Entry) error {
	fmt.Printf("[%s] %s\n", e.Level, e.Message)
	return nil
}
