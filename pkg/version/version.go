package version

var (
	gitVersion = "0.0.0"                // `git describe --tags --long --dirty` or "no git" if not set
	gitCommit  = "no git"               // `git rev-parse HEAD` or "no git" if not set
	buildDate  = "1970-01-01T00:00:00Z" // `date -u +'%Y-%m-%dT%H:%M:%SZ'`, ISO8601 format
)

// BuildInfo is a struct which contains (version, commit, date) for aptomi binary, so that we know when and how it was built
type BuildInfo struct {
	GitVersion string
	GitCommit  string
	BuildDate  string
}

// GetBuildInfo returns BuildInfo for aptomi
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		gitVersion,
		gitCommit,
		buildDate,
	}
}

func (buildInfo *BuildInfo) GetDefaultColumns() []string {
	return []string{"Git Version", "Git Commit", "Build Date"}
}

func (buildInfo *BuildInfo) AsColumns() map[string]string {
	result := make(map[string]string)

	result["Git Version"] = buildInfo.GitVersion
	result["Git Commit"] = buildInfo.GitCommit
	result["Build Date"] = buildInfo.BuildDate

	return result
}

/*
// GetDefaultColumns returns default set of columns to be displayed
func (policyData *PolicyData) GetDefaultColumns() []string {
	return []string{"Policy Version"}
}

// AsColumns returns PolicyData representation as columns
func (policyData *PolicyData) AsColumns() map[string]string {
	result := make(map[string]string)

	result["Policy Version"] = policyData.GetGeneration().String()

	return result
}
*/
