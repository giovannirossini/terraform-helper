package version

import "fmt"

// These variables are set via build flags. See Makefile for details.
var (
	// Version is the version of the application.
	Version = "dev"
	// Commit is the git commit hash.
	Commit = "unknown"
	// BuildTime is the build timestamp.
	BuildTime = "unknown"
)

// String returns a formatted version string.
func String() string {
	return fmt.Sprintf("terraform-helper version %s (commit: %s, built: %s)", Version, Commit, BuildTime)
}
