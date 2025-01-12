module github.com/woozymasta/a2s

go 1.23.1

require (
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.27.5
	github.com/woozymasta/steam v0.1.2
	internal/vars v0.0.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace internal/vars => ./internal/vars
