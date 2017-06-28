package log

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
)

// PlainMemoryLogger represents a logger that logsinto memory
type PlainMemoryLogger struct {
	buf    *bytes.Buffer
	logger *log.Logger
}

// NewPlainMemoryLogger creates a new PlainMemoryLogger
func NewPlainMemoryLogger(verbose bool) PlainMemoryLogger {
	buf := &bytes.Buffer{}
	logger := log.New()
	logger.Out = buf
	logger.Formatter = new(PlainFormatter)
	if verbose {
		logger.Level = log.DebugLevel
	} else {
		logger.Level = log.InfoLevel
	}
	return PlainMemoryLogger{buf: buf, logger: logger}
}

// GetLogger returns the underlying logger
func (pml *PlainMemoryLogger) GetLogger() *log.Logger {
	return pml.logger
}

// GetBuffer returns the underlying memory buffer
func (pml *PlainMemoryLogger) GetBuffer() *bytes.Buffer {
	return pml.buf
}
