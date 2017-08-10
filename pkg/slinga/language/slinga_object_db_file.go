package language

import (
	"github.com/mattn/go-zglob"
	"sort"
	"path/filepath"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	"strings"
)

type SlingaObjectDatabaseDir struct {
	dir string
}

func NewSlingaObjectDatabaseDir(dir string) SlingaObjectDatabase {
	return &SlingaObjectDatabaseDir{dir: dir}
}

func (db *SlingaObjectDatabaseDir) LoadPolicyObjects(revision int, namespace string) *Policy {
	files, _ := zglob.Glob(filepath.Join(db.dir, "**", "*.yaml"))
	sort.Strings(files)

	policy := NewPolicy()
	parser := NewSlingaObjectParser()
	for _, f := range files {
		if !strings.Contains(f, "external") {
			objects := *yaml.LoadObjectFromFile(f, new([]*SlingaObject)).(*[]*SlingaObject)
			for _, obj := range objects {
				policy.addObject(parser.parseObject(obj))
			}
		}
	}

	return policy
}
