package slinga

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
)

// Tracing logger is for detailed messages printed via --verbose
var trace *TraceLogger

// Debug logger writes debug information into a file
var debug *log.Logger

// TraceLogger is a logger that prints onto the screen and supports on/off
type TraceLogger struct {
	logger *log.Logger
}

// PlainFormatter just formats messages into plain text
type PlainFormatter struct{}

// Format just returns entry message without formatting it
func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

func (screenLogger *TraceLogger) setEnable(enabled bool) {
	if enabled {
		trace.logger.Level = log.DebugLevel
	} else {
		trace.logger.Level = log.PanicLevel
	}
}

// Printf prints information onto screen
func (screenLogger *TraceLogger) Printf(depth int, format string, args ...interface{}) {
	format = strings.Repeat(" ", 4*depth) + format + "\n"
	screenLogger.logger.Printf(format, args...)
}

// Println prints new line onto screen
func (screenLogger *TraceLogger) Println() {
	screenLogger.logger.Println()
}

// SetDebugLevel sets level for the debug logger
func SetDebugLevel(level log.Level) {
	debug.Level = level
}

func init() {
	trace = &TraceLogger{
		logger: log.New(),
	}

	debug = log.New()

	// if we are running normally (not in unit tests)
	if flag.Lookup("test.v") == nil {
		// Make sure we have a place to write the current state to
		// This is a bit of a hack to clean up the current directory here
		// But you can't really do this in policy.go, because this will cause a race condition with init()
		PrepareCurrentRunDirectory(GetAptomiBaseDir())

		// setup debug logger
		// don't log much by default. log level will be overridden when called from CLU
		fileNameDebug := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypeLogs, "debug.log")
		debug.Out, _ = os.OpenFile(fileNameDebug, os.O_CREATE|os.O_WRONLY, 0644)
		debug.Level = log.PanicLevel

		// setup tracing logger
		// don't log much by default. log level will be overridden when called from CLU
		fileNameTrace := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypeLogs, "trace.log")
		trace.logger.Out, _ = os.OpenFile(fileNameTrace, os.O_CREATE|os.O_WRONLY, 0644)
		trace.logger.Formatter = new(PlainFormatter)
		trace.logger.Level = log.PanicLevel

		// Add a hook to print important errors to stdout as well
		debug.Hooks.Add(&logHook{})
	} else {
		// running under unit tests
		debug.Level = log.WarnLevel
		trace.logger.Level = log.WarnLevel
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
