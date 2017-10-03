package event

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

type HookStdout struct {
}

func (buf *HookStdout) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (buf *HookStdout) Fire(e *logrus.Entry) error {
	fmt.Printf("[%s] %s\n", e.Level, e.Message)
	return nil
}
