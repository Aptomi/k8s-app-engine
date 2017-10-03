package event

import (
	"github.com/Sirupsen/logrus"
	"strings"
)

type UnitTestLogVerifier struct {
	checkForErrorMessage string
	cnt                  int
}

func NewUnitTestLogVerifier(checkForErrorMessage string) *UnitTestLogVerifier {
	return &UnitTestLogVerifier{checkForErrorMessage: checkForErrorMessage}
}

func (verifier *UnitTestLogVerifier) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (verifier *UnitTestLogVerifier) Fire(e *logrus.Entry) error {
	if e.Level == logrus.ErrorLevel && strings.Contains(e.Message, verifier.checkForErrorMessage) {
		verifier.cnt++
	}
	return nil
}

func (verifier *UnitTestLogVerifier) MatchedErrorsCount() int {
	return verifier.cnt
}
