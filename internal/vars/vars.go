// Package vars provides build-time metadata about the application.
// Values are typically injected at build time via ldflags and reflect
// the state of the git repository and the build moment.
package vars

import (
	"fmt"
	"os"
	"time"
)

var (
	// Version is the application version (usually a git tag or semver),
	// defaults to "dev".
	Version = "dev"

	// Commit is the current git commit SHA (short or full), defaults to "unknown".
	Commit = "unknown"

	// BuildTime is the application build time in RFC3339 UTC, defaults to 1970-01-01.
	BuildTime = time.Unix(0, 0)

	// URL is the repository URL (https), defaults to "https://github.com/woozymasta/a2s".
	URL = "https://github.com/woozymasta/a2s"

	// _buildTime is an internal string passed via ldflags that overrides BuildTime when set.
	_buildTime string
)

// BuildInfo is a safe container for build metadata that can be
// exposed externally (e.g. via an API or CLI command).
type BuildInfo struct {
	// Version is the application version (usually a git tag or semver).
	Version string `json:"version"`

	// Commit is the current git commit SHA (short or full).
	Commit string `json:"commit"`

	// BuildTime is the application build time (UTC).
	BuildTime time.Time `json:"build_time,omitempty"`

	// URL is the repository URL.
	URL string `json:"url,omitempty"`
}

func init() {
	if _buildTime != "" {
		if t, err := time.Parse(time.RFC3339, _buildTime); err == nil {
			BuildTime = t.UTC()
		}
	}
}

// Print writes build information (URL, binary file path, version,
// commit and build time) to standard output in a human-readable format.
func Print() {
	fmt.Printf(`url:      %s
file:     %s
version:  %s
commit:   %s
built:    %s
`, URL, os.Args[0], Version, Commit, BuildTime)
}

// Info returns a BuildInfo struct populated with the current
// build metadata values.
func Info() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		URL:       URL,
	}
}
