package db

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
)

// LastRevisionAbsentValue represents initial revision ("zero revision")
const LastRevisionAbsentValue = 0

// AptomiRevision represents revision
type AptomiRevision int

// AptomiRun represents data about the aptomi run
type AptomiRun struct {
	// Revision of configuration
	Revision AptomiRevision

	// Metadata of the run (e.g. commit id, start time, end time, etc)
	Metadata util.NestedParameterMap
}

// Increment increments the revision
func (revision AptomiRevision) Increment() AptomiRevision {
	return AptomiRevision(revision + 1)
}

// String returns aptomi revision converted to string
func (revision AptomiRevision) String() string {
	if revision <= LastRevisionAbsentValue {
		return "N/A"
	}
	return fmt.Sprintf("%d", revision)
}
