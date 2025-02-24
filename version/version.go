package version

const (
	// CoreSemVer represents the core semantic version when not using git describe.
	// It follows the semantic versioning format.
	CoreSemVer = "0.1.1"
)

// GitCommitHash uses git rev-parse HEAD to find commit hash which is helpful
// for the engineering team when working with the pell emulator binary. See Makefile
var GitCommitHash = ""
