package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
)

// SummaryView represents summary view that we show on the home page
type SummaryView struct {
	userID string
	state  slinga.ServiceUsageState
}

// NewObjectView creates a new SummaryView
func NewSummaryView(userID string, state slinga.ServiceUsageState) SummaryView {
	return SummaryView{
		userID: userID,
		state:  state,
	}
}

// GetData returns data for a given view
func (view SummaryView) GetData() interface{} {
	return view.userID
}
