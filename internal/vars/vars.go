// Package vars is an internal technical variable store used at build time,
// populated with values ​​based on the state of the git repository.
package vars

var (
	// Version of application (git tag) semver/tag, e.g. v1.2.3
	Version string = "dev"

	// Current git commit, full or short git SHA
	Commit string = "unknown"

	// Time of start build app, RFC3339 UTC
	BuildTime string = "1970-01-01T00:00:00Z"

	// URL to repository (https)
	URL string = "https://github.com/woozymasta/a2s"
)
