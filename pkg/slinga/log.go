package slinga

import (
	"github.com/Sirupsen/logrus"
	"fmt"
	"os"
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

func (logger *ScreenLogger) Printf(depth int, format string, args ...interface{}) {
	if logger.enabled {
		indent := ""
		for n := 0; n <= 4*depth; n++ {
			indent = indent + " "
		}
		format = indent + format + "\n"
		fmt.Printf(format, args...)
	}
}

func (logger *ScreenLogger) Println() {
	if logger.enabled {
		fmt.Println()
	}
}

func SetDebugLevel(level logrus.Level) {
	debug.Level = level
}

func init() {
	tracing = &ScreenLogger{}

	debug = logrus.New()
	debug.Out, _ = os.OpenFile(GetAptomiDBDir() + "/" + "debug.log", os.O_CREATE|os.O_WRONLY, 0644)
	SetDebugLevel(logrus.ErrorLevel)
}
