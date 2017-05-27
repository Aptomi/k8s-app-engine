package slinga

import (
	"github.com/Sirupsen/logrus"
	"log"
)

var tracing *ScreenLogger
var debug *logrus.Logger

// ScreenLogger contains is a logger that prints onto the screen and supports on/off
type ScreenLogger struct {
	enabled  bool
}

func (logger *ScreenLogger) setEnable(enabled bool) {
	logger.enabled = enabled
}

func (logger *ScreenLogger) log(depth int, format string, args ...interface{}) {
	if logger.enabled {
		indent := ""
		for n := 0; n <= 4*depth; n++ {
			indent = indent + " "
		}
		format = indent + format
		log.Printf(format, args...)
	}
}

func (logger *ScreenLogger) newline() {
	logger.log(0, "\n")
}

func SetDebugLevel(level logrus.Level) {
	debug.Level = level
}

func init() {
	tracing = &ScreenLogger{}

	debug = logrus.New()
	SetDebugLevel(logrus.ErrorLevel)
}
