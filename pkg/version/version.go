package version

import "github.com/Aptomi/aptomi/pkg/runtime"

var (
	gitVersion = "0.0.0"                // `git describe --tags --long --dirty` or "no git" if not set
	gitCommit  = "no git"               // `git rev-parse HEAD` or "no git" if not set
	buildDate  = "1970-01-01T00:00:00Z" // `date -u +'%Y-%m-%dT%H:%M:%SZ'`, ISO8601 format
)

// GetBuildInfo returns BuildInfo for aptomi
func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		TypeKind:   BuildInfoObject.GetTypeKind(),
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
	}
}

// BuildInfoObject is an informational data structure with Kind and Constructor for Version
var BuildInfoObject = &runtime.TypeInfo{
	Kind:        "version",
	Constructor: func() runtime.Object { return &BuildInfo{} },
}

// BuildInfo represents version, commit and date for aptomi binary, so that we know when and how it was built
type BuildInfo struct {
	runtime.TypeKind `yaml:",inline"`
	GitVersion       string
	GitCommit        string
	BuildDate        string
}

// GetDefaultColumns returns default set of columns to be displayed
func (buildInfo *BuildInfo) GetDefaultColumns() []string {
	return []string{"Git Version", "Git Commit", "Build Date"}
}

// AsColumns returns PolicyData representation as columns
func (buildInfo *BuildInfo) AsColumns() map[string]string {
	result := make(map[string]string)

	result["Git Version"] = buildInfo.GitVersion
	result["Git Commit"] = buildInfo.GitCommit
	result["Build Date"] = buildInfo.BuildDate

	return result
}
