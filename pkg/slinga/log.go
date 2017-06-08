package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"flag"
)

// Tracing logger is for detailed messages printed via --verbose
var tracing *ScreenLogger

// Debug logger writes debug information into a file
var debug *log.Logger

// ScreenLogger is a logger that prints onto the screen and supports on/off
type ScreenLogger struct {
	enabled bool
}

func (logger *ScreenLogger) setEnable(enabled bool) {
	logger.enabled = enabled
}

// Printf prints information onto screen
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

// Println prints new line onto screen
func (logger *ScreenLogger) Println() {
	if logger.enabled {
		fmt.Println()
	}
}

// SetDebugLevel sets level for the debug logger
func SetDebugLevel(level log.Level) {
	debug.Level = level
}

func init() {
	tracing = &ScreenLogger{}

	debug = log.New()

	if flag.Lookup("test.v") == nil {
		// running normally
		debug.Out, _ = os.OpenFile(GetAptomiDBDir()+"/"+"debug.log", os.O_CREATE|os.O_WRONLY, 0644)

		// Don't log much by default. It will be overridden with "--debug" from CLI
		debug.Level = log.PanicLevel

		// Add a hook to print important errors to stdout as well
		debug.Hooks.Add(&logHook{})
	} else {
		// running under unit tests
		debug.Level = log.WarnLevel
	}
}

type logHook struct {
}

func (l *logHook) Levels() []log.Level {
	return []log.Level{
		log.WarnLevel,
		log.ErrorLevel,
		log.FatalLevel,
		log.PanicLevel,
	}
}

func (l *logHook) Fire(e *log.Entry) error {
	fmt.Println("Error!")
	fmt.Printf("  %s\n", e.Message)
	fmt.Printf("  %v\n", e.Data)
	return nil
}
