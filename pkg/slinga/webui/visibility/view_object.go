package visibility

import (
	log "github.com/Sirupsen/logrus"
)

// ObjectView represents an in-depth view for a particular object
type ObjectView struct {
	id string
	//revision *resolve.Revision
}

// NewObjectView creates a new ObjectView
func NewObjectView(id string) ObjectView {
	return ObjectView{
		id: id,
		//revision: resolve.LoadRevision(),
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

	return obj.getDetails(obj.isItMyID(ov.id) /*, ov.revision*/)
}
