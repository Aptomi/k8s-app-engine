package db

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
	log "github.com/Sirupsen/logrus"
)

// ResetAptomiState fully resets aptomi state by deleting all files and directories from its database
// That includes all revisions of policy, resolution data, logs, etc
func ResetAptomiState() {
	baseDir := GetAptomiBaseDir()
	Debug.WithFields(log.Fields{
		"baseDir": baseDir,
	}).Info("Resetting aptomi state")

	err := DeleteDirectoryContents(baseDir)
	if err != nil {
		Debug.WithFields(log.Fields{
			"directory": baseDir,
			"error":     err,
		}).Panic("Directory contents can't be deleted")
	}

	fmt.Println("Aptomi state is now empty. Deleted all objects")
}
