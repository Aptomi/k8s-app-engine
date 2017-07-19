package db

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/mattn/go-zglob"
	"path/filepath"
)

// GetLastRevision returns the last revision as integer
func GetLastRevision(baseDir string) AptomiRevision {
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeRevision))
	if len(files) <= 0 {
		// If there are no revision files, return first revision
		return LastRevisionAbsentValue
	}
	if len(files) > 1 {
		panic("Found more than one revision files")
	}
	return loadRevisionFromFile(files[0])
}

// Loads revision from file
func loadRevisionFromFile(fileName string) AptomiRevision {
	return *yaml.LoadObjectFromFile(fileName, new(AptomiRevision)).(*AptomiRevision)
}

// SaveAsLastRevision stores last revision in a file under Aptomi DB
func (revision AptomiRevision) SaveAsLastRevision() {
	fileName := GetAptomiObjectWriteFileGlobal(GetAptomiBaseDir(), TypeRevision)
	yaml.SaveObjectToFile(fileName, revision)
}

// SaveCurrentRun saves contents of the current run
func (revision AptomiRevision) SaveCurrentRun() {
	// where run is saved to
	var err error
	savedRunDir := filepath.Join(GetAptomiBaseDir(), revision.GetRunDirectory())

	// copy over current run directory
	currentRunDir := filepath.Join(GetAptomiBaseDir(), AptomiCurrentRunDir)
	err = CopyDirectory(currentRunDir, savedRunDir)
	if err != nil {
		panic(err)
	}

	// copy over logs directory
	logsDir := filepath.Join(GetAptomiBaseDir(), string(TypeLogs))
	err = CopyDirectory(logsDir, filepath.Join(savedRunDir, string(TypeLogs)))
	if err != nil {
		panic(err)
	}

	// TODO: save metadata about the run
}

// GetRunDirectory returns run directory, formatted as string
func (revision AptomiRevision) GetRunDirectory() string {
	return fmt.Sprintf("run-%09d", revision)
}
