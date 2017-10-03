package event

import (
	"github.com/Sirupsen/logrus"
)

type messageBuffer struct {
	entries []*logrus.Entry
}

func (buf *messageBuffer) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (buf *messageBuffer) Fire(e *logrus.Entry) error {
	buf.entries = append(buf.entries, e)
	return nil
}
