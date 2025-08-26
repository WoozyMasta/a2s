package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/a2s/internal/ping"
)

// Options defines CLI options for controlling ping behavior and output formatting.
type Options struct {
	// AppID for more accurate results
	AppID uint64 `short:"i" long:"app-id"`

	// Timeout specifies the timeout in seconds for each request.
	Timeout int `short:"t" long:"deadline-timeout" default:"3"`

	// PingCount specifies the number of ping requests to send.
	// A value of 0 means sending requests indefinitely until interrupted.
	PingCount int `short:"c" long:"ping-count" default:"0"`

	// PingPeriod specifies the delay in seconds between consecutive pings.
	PingPeriod int `short:"p" long:"ping-period" default:"1"`

	// Buffer specifies the buffer size in bytes for network operations.
	Buffer uint16 `short:"b" long:"buffer-size" default:"8096"`

	// JSON enables output formatting in JSON.
	JSON bool `short:"j" long:"json"`

	// Help shows usage information and available command-line options.
	Help bool `short:"h" long:"help"`

	// Version shows version and build information.
	Version bool `short:"v" long:"version"`
}

func main() {
	opts := &Options{}
	p := flags.NewParser(opts, flags.PassDoubleDash)
	args, err := p.Parse()
	if err != nil {
		fatal(err)
	}

	if opts.Help {
		printHelp(true)
	}
	if opts.Version {
		printVersion()
	}

	if len(args) < 1 {
		printHelp(false)
		fatal("Command must be provided")
	}
	if len(args) < 2 {
		fatal("Host and port must be provided as positional arguments")
	}
	if len(args) > 3 {
		fatalf("Extra command passed %s", args[3:])
	}

	host := args[1]
	if len(args) > 2 {
		host += ":" + args[2]
	}

	client := createClient(host, opts.Timeout, opts.Buffer)

	switch args[0] {
	case "info":
		printInfo(client, opts.JSON)
	case "rules":
		printRules(client, opts.AppID, opts.JSON)
	case "players":
		printPlayers(client, opts.JSON)
	case "all":
		printInfo(client, opts.JSON)
		printRules(client, opts.AppID, opts.JSON)
		printPlayers(client, opts.JSON)
	case "ping":
		ping.Start(client, opts.PingCount, opts.PingPeriod)
	default:
		fatalf("Unknown command '%s'", args[0])
	}

	if err := client.Close(); err != nil {
		fatal("Cant close connection")
	}
}
