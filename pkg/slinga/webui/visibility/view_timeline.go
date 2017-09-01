package visibility

import (
	"sort"
)

// TimelineView represents timeline view
type TimelineView struct {
	userID string
	// revisions map[int]*resolve.Revision
}

// NewTimelineView creates a new TimelineView
func NewTimelineView(userID string) TimelineView {
	return TimelineView{
		userID: userID,
		// revisions: resolve.LoadRevisionsAll(),
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
	/*
		for revisionNumber, revision := range view.revisions {
			rev := AptomiRevision(revisionNumber)
			entry := lineEntry{
				"id":             rev.GetRunDirectory(),
				"revisionNumber": strconv.Itoa(revisionNumber),
				"dir":            rev.GetRunDirectory(),
				"createdOn":      revision.Resolution.CreatedOn,
				"diff":           "todo: WE NEED TO THINK WHERE TO STORE PLAIN-TEXT-DIFF (IT USED TO BE STORED IN A WRONG PLACE)",
			}
			result = append(result, entry)
		}
	*/
	sort.Sort(sort.Reverse(result))
	return result
}
