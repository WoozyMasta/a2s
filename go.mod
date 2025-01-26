module github.com/woozymasta/a2s

go 1.23.1

require (
	github.com/jessevdk/go-flags v1.6.1
	github.com/woozymasta/steam v0.1.3
	internal/vars v0.0.0
)

require golang.org/x/sys v0.29.0 // indirect

replace (
	internal/bread => ./internal/bread
	internal/ping => ./internal/ping
	internal/tableprinter => ./internal/tableprinter
	internal/vars => ./internal/vars
)
