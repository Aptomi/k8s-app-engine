package event

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

// HookConsole implements event log hook, which prints all event log entries to the console (stdout)
type HookConsole struct {
}

// Levels defines on which log levels this hook should be fired
func (buf *HookConsole) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (buf *HookConsole) Fire(e *logrus.Entry) error {
	msg := e.Message
	if scope, ok := e.Data["scope"]; ok {
		msg = "(" + scope.(string) + ") " + msg
	}

	fmt.Printf("[%s] %s\n", e.Level, msg)
	return nil
}
