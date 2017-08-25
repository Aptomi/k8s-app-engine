package eventlog

import (
	"github.com/Sirupsen/logrus"
	"fmt"
)

type HookBoltDB struct {
}

func (buf *HookBoltDB) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (buf *HookBoltDB) Fire(e *logrus.Entry) error {
	// figure out to which objects this entry should be attached to
	attachedToObjects := e.Data["attachedTo"].(*AttachedObjects).objects
	delete(e.Data, "attachedTo")

	// TODO: store this entry into bolt
	fmt.Printf("[%s] %s %p %p\n", e.Level, e.Message, e.Data, attachedToObjects)
	return nil
}
