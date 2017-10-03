package event

import (
	"github.com/Sirupsen/logrus"
	"strings"
)

// UnitTestLogVerifier is a mock logger and a unit test helper for verifying event log messages
type UnitTestLogVerifier struct {
	checkForErrorMessage string
	cnt                  int
}

// NewUnitTestLogVerifier creates a new UnitTestLogVerifier which searches for a given error message
func NewUnitTestLogVerifier(checkForErrorMessage string) *UnitTestLogVerifier {
	return &UnitTestLogVerifier{checkForErrorMessage: checkForErrorMessage}
}

// Levels returns a set of levels for the mock logger. Returns all levels
func (verifier *UnitTestLogVerifier) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a log entry. If it contains a given error message, it increments verifier.cnt
func (verifier *UnitTestLogVerifier) Fire(e *logrus.Entry) error {
	if e.Level == logrus.ErrorLevel && strings.Contains(e.Message, verifier.checkForErrorMessage) {
		verifier.cnt++
	}
	return nil
}

// MatchedErrorsCount returns verifier.cnt, which represent the number of found errors matching a given string
func (verifier *UnitTestLogVerifier) MatchedErrorsCount() int {
	return verifier.cnt
}
