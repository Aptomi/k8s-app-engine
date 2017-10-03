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
