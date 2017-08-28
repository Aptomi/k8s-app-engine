package visibility

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	"sort"
	"strconv"
)

// TimelineView represents timeline view
type TimelineView struct {
	userID string
	states map[int]engine.ServiceUsageState
}

// NewTimelineView creates a new TimelineView
func NewTimelineView(userID string, states map[int]engine.ServiceUsageState) TimelineView {
	return TimelineView{
		userID: userID,
		states: states,
	}
}

// GetData returns data for a given view
func (view TimelineView) GetData() interface{} {
	result := lineEntryList{}
	/*
		if !view.users.Users[view.userID].IsGlobalOps() {
			return result
		}
	*/
	for revisionNumber, state := range view.states {
		rev := AptomiRevision(revisionNumber)
		entry := lineEntry{
			"id":             rev.GetRunDirectory(),
			"revisionNumber": strconv.Itoa(revisionNumber),
			"dir":            rev.GetRunDirectory(),
			"createdOn":      state.CreatedOn,
			"diff":           state.DiffAsText,
		}
		result = append(result, entry)
	}
	sort.Sort(sort.Reverse(result))
	return result
}
