package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// If revision is absent, assume this value
const lastRevisionAbsentValue = 0

// AptomiRevision represents revision
type AptomiRevision int

// AptomiRun represents data about the aptomi run
type AptomiRun struct {
	// Revision of configuration
	Revision AptomiRevision

	// Metadata of the run (e.g. commit id, start time, end time, etc)
	Metadata NestedParameterMap
}

// Increments a revision
func (revision AptomiRevision) increment() AptomiRevision {
	return AptomiRevision(revision + 1)
}

// GetLastRevision returns the last revision as integer
func GetLastRevision(baseDir string) AptomiRevision {
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeRevision))
	if len(files) <= 0 {
		// If there are no revision files, return first revision
		return lastRevisionAbsentValue
	}
	if len(files) > 1 {
		debug.WithFields(log.Fields{
			"files": files,
		}).Fatal("Found more than one revision files")
	}
	return loadRevisionFromFile(files[0])
}

func loadRevisionFromFile(fileName string) AptomiRevision {
	dat, e := ioutil.ReadFile(fileName)

	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}

	var t = AptomiRevision(lastRevisionAbsentValue)
	e = yaml.Unmarshal([]byte(dat), &t)

	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal revision")
	}

	return t
}

// SaveLastRevision stores last revision in a file under Aptomi DB
func (revision AptomiRevision) saveAsLastRevision() {
	fileName := GetAptomiObjectWriteFileGlobal(GetAptomiBaseDir(), TypeRevision)

	debug.WithFields(log.Fields{
		"file": fileName,
	}).Info("Saving last revision")

	e := ioutil.WriteFile(fileName, []byte(serializeObject(revision)), 0644)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to save revision")
	}
}

// Saves contents of the current run
func (revision AptomiRevision) saveCurrentRun() {
	currentRunDir := filepath.Join(GetAptomiBaseDir(), aptomiCurrentRunDir)
	savedRunDir := filepath.Join(GetAptomiBaseDir(), revision.getRunDirectory())
	copyDirectory(currentRunDir, savedRunDir)

	// TODO: save data about the run
}

func (revision AptomiRevision) getRunDirectory() string {
	return fmt.Sprintf("run-%09d", revision)
}

func (revision AptomiRevision) String() string {
	if revision <= lastRevisionAbsentValue {
		return "N/A"
	}
	return fmt.Sprintf("%d", revision)
}
