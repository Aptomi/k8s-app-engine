package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRevision(t *testing.T) {
	revision := GetLastRevision("testdata/unittests")
	assert.Equal(t, AptomiRevision(239), revision, "Correct revision expected")
}
