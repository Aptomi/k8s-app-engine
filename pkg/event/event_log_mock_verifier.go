package event

import (
	"github.com/Sirupsen/logrus"
	"strings"
)

// LogVerifier is a mock logger and a unit test helper for verifying event log messages
type LogVerifier struct {
	expectedMessage string
	isError         bool
	cnt             int
}

// NewLogVerifier creates a new LogVerifier which searches for a given error message
func NewLogVerifier(expectedMessage string, isError bool) *LogVerifier {
	return &LogVerifier{expectedMessage: expectedMessage, isError: isError}
}

// Levels returns a set of levels for the mock logger. Returns all levels
func (verifier *LogVerifier) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a log entry. If it contains a given error message, it increments verifier.cnt
func (verifier *LogVerifier) Fire(e *logrus.Entry) error {
	if len(verifier.expectedMessage) > 0 && strings.Contains(e.Message, verifier.expectedMessage) {
		if !verifier.isError || (verifier.isError && e.Level == logrus.ErrorLevel) {
			verifier.cnt++
		}
	}
	return nil
}

// MatchedErrorsCount returns verifier.cnt, which represent the number of found errors matching a given string
func (verifier *LogVerifier) MatchedErrorsCount() int {
	return verifier.cnt
}
