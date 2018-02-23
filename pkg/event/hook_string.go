package event

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
)

// HookString saves all events into the single string
type HookString struct {
	buf bytes.Buffer
}

// Levels defines on which log levels this hook should be fired
func (hook *HookString) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (hook *HookString) Fire(e *logrus.Entry) error {
	_, err := hook.buf.WriteString(fmt.Sprintf("[%s] %s\n", e.Level, e.Message))
	return err
}

// String returns string representation of all buffered events
func (hook *HookString) String() string {
	return hook.buf.String()
}
