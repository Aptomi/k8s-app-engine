package db

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

// ResetAptomiState fully resets aptomi state by deleting all files and directories from its database
// That includes all revisions of policy, resolution data, logs, etc
func ResetAptomiState() {
	baseDir := GetAptomiBaseDir()

	err := DeleteDirectoryContents(baseDir)
	if err != nil {
		panic(fmt.Sprintf("Directory '%s' contents can't be deleted: %s", baseDir, err.Error()))
	}

	fmt.Println("Aptomi state is now empty. Deleted all objects")
}
