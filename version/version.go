package version

import (
	"fmt"
	"runtime"
)

var (
	gitRepo      = ""
	gitTag       = ""
	gitCommit    = "none"           // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState = "not a git tree" // state of git tree, either "clean" or "dirty"
	buildDate    = "unknown"        // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

// Info contains versioning information.
type Info struct {
	GitRepo      string `json:"gitRepo"`
	GitTag       string `json:"gitTag"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// String returns info as a human-friendly version string.
// Version is the SHA of the git commit from which this binary was built.
func (info *Info) String() string {
	return fmt.Sprintf("%s - %s", info.GitTag, info.GitCommit)
}

// Get 返回详细的版本信息
func Get() Info {
	return Info{
		GitRepo:      gitRepo,
		GitTag:       gitTag,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
