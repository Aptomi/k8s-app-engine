package event

import (
	"github.com/Sirupsen/logrus"
)

// HookLogger is a hook for logrus that prints messages to console using logger
type HookLogger struct {
}

// Levels defines on which log levels this hook should be fired
func (hook *HookLogger) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (hook *HookLogger) Fire(entry *logrus.Entry) error {
	msg := entry.Message
	if scope, ok := entry.Data["scope"]; ok {
		msg = "(" + scope.(string) + ") " + msg
	}

	switch entry.Level {
	case logrus.PanicLevel:
		logrus.Panic(msg)
	case logrus.FatalLevel:
		logrus.Fatal(msg)
	case logrus.ErrorLevel:
		logrus.Error(msg)
	case logrus.WarnLevel:
		logrus.Warn(msg)
	case logrus.InfoLevel:
		logrus.Info(msg)
	case logrus.DebugLevel:
		logrus.Debug(msg)
	}

	return nil
}
