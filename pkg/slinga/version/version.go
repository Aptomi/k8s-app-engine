package version

var (
	gitVersion string = "0.0.0"                // `git describe --tags --long --dirty` or "no git" if not set
	gitCommit  string = "no git"               // `git rev-parse HEAD` or "no git" if not set
	buildDate  string = "1970-01-01T00:00:00Z" // `date -u +'%Y-%m-%dT%H:%M:%SZ'`, ISO8601 format
)

type BuildInfo struct {
	GitVersion string
	GitCommit  string
	BuildDate  string
}

func GetBuildInfo() BuildInfo {
	return BuildInfo{
		gitVersion,
		gitCommit,
		buildDate,
	}
}
