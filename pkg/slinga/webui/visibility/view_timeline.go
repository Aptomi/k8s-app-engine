package visibility

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"sort"
	"strconv"
)

// TimelineView represents timeline view
type TimelineView struct {
	userID string
	states map[int]*resolve.ResolvedState
}

// NewTimelineView creates a new TimelineView
func NewTimelineView(userID string) TimelineView {
	return TimelineView{
		userID: userID,
		states: resolve.LoadResolvedStatesAll(),
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
			"createdOn":      state.State.CreatedOn,
			"diff":           "WE NEED TO THINK WHERE TO STORE PLAIN-TEXT-DIFF (IT USED TO BE STORED IN A WRONG PLACE)",
		}
		result = append(result, entry)
	}
	sort.Sort(sort.Reverse(result))
	return result
}
