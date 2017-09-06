package actual

import "github.com/Aptomi/aptomi/pkg/slinga/object"

type StateUpdater interface {
	Create(obj object.Base) error
	Update(obj object.Base) error
	Delete(string) error
}
