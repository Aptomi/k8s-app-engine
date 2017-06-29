package log

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
)

// Debug logger writes debug information into a file
var Debug *log.Logger

// SetDebugLevel sets level for the debug logger
func SetDebugLevel(level log.Level) {
	Debug.Level = level
}

// SetLogFileName redirects log to a file
func SetLogFileName(fileName string) {
	// Redirect to a file
	Debug.Out, _ = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)

	// Add a hook to print important errors to stdout as well
	Debug.Hooks.Add(&logHook{})
}

func init() {
	Debug = log.New()

	// if we are running normally (not in unit tests)
	if flag.Lookup("test.v") == nil {
		// don't log much by default. log level will be overridden when called from CLI
		Debug.Level = log.PanicLevel
	} else {
		// running under unit tests
		Debug.Level = log.WarnLevel
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
