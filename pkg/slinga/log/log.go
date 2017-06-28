package log

import (
	"flag"
	"fmt"
	. "github.com/Frostman/aptomi/pkg/slinga/fileio"
	log "github.com/Sirupsen/logrus"
	"os"
)

// Debug logger writes debug information into a file
var Debug *log.Logger

// SetDebugLevel sets level for the debug logger
func SetDebugLevel(level log.Level) {
	Debug.Level = level
}

func init() {
	Debug = log.New()

	// if we are running normally (not in unit tests)
	if flag.Lookup("test.v") == nil {
		// Make sure we have a clean place to write the current run to
		// This is a bit of a hack to clean up the current directory here
		// But you can't really do this in policy.go, because this will cause a race condition with init()
		PrepareCurrentRunDirectory(GetAptomiBaseDir())

		// setup debug logger
		// don't log much by default. log level will be overridden when called from CLU
		fileNameDebug := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypeLogs, "debug.log")
		Debug.Out, _ = os.OpenFile(fileNameDebug, os.O_CREATE|os.O_WRONLY, 0644)
		Debug.Level = log.PanicLevel

		// Add a hook to print important errors to stdout as well
		Debug.Hooks.Add(&logHook{})
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
