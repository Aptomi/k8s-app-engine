package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRevision(t *testing.T) {
	revision := GetLastRevision("../testdata/unittests_new")
	assert.Equal(t, AptomiRevision(239), revision, "Correct revision expected")

	revisionNonExistent := GetLastRevision("../testdata/unittests_new/non-existent")
	assert.Equal(t, AptomiRevision(LastRevisionAbsentValue), revisionNonExistent, "Correct initial revision expected")
}
