package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga/engine"
	log "github.com/Sirupsen/logrus"
)

// ObjectView represents an in-depth view for a particular object
type ObjectView struct {
	id    string
	state engine.ServiceUsageState
}

// NewObjectView creates a new ObjectView
func NewObjectView(id string, state engine.ServiceUsageState) ObjectView {
	return ObjectView{
		id:    id,
		state: state,
	}
}

// GetData returns graph for a given view
func (ov ObjectView) GetData() interface{} {
	obj := getLoadableObject(ov.id)
	if obj == nil {
		log.WithFields(log.Fields{
			"id": ov.id,
		}).Warning("Unable to load object")
		return nil
	}

	return obj.getDetails(obj.isItMyID(ov.id), ov.state)
}
