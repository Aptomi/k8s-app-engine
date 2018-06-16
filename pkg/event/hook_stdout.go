package event

import (
	"os"

	"github.com/sirupsen/logrus"
)

// HookConsole implements event log hook, which prints entries to the console
type HookConsole struct {
	logger *logrus.Logger
}

// NewHookConsole creates a new HookConsole
func NewHookConsole(level logrus.Level) *HookConsole {
	return &HookConsole{
		logger: &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     level,
		},
	}
}

// Levels defines on which log levels this hook should be fired
func (hook *HookConsole) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (hook *HookConsole) Fire(e *logrus.Entry) error {
	msg := e.Message
	if scope, ok := e.Data["scope"]; ok {
		msg = "(" + scope.(string) + ") " + msg
	}

	switch e.Level {
	case logrus.PanicLevel:
		hook.logger.Panic(msg)
	case logrus.FatalLevel:
		hook.logger.Fatal(msg)
	case logrus.ErrorLevel:
		hook.logger.Error(msg)
	case logrus.WarnLevel:
		hook.logger.Warn(msg)
	case logrus.InfoLevel:
		hook.logger.Info(msg)
	case logrus.DebugLevel:
		hook.logger.Debug(msg)
	}

	return nil
}
