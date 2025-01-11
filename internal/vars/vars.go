// Package vars is an internal technical variable store used at build time,
// populated with values ​​based on the state of the git repository.
package vars

var (
	Version   string // Version of application (git tag)
	Commit    string // Current git commit
	BuildTime string // Time of start build app
	URL       string // URL to repository
)
