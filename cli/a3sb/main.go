package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/a2s/internal/ping"
)

// CLI options
type Options struct {
	AppID      uint64 `short:"i" long:"app-id"`                       // AppID for more accurate results
	Timeout    int    `short:"t" long:"deadline-timeout" default:"3"` // Set timeout in seconds
	PingCount  int    `short:"c" long:"ping-count" default:"0"`       // Set the number of ping requests to send
	PingPeriod int    `short:"p" long:"ping-period" default:"1"`      // Set the period between pings in seconds
	Buffer     uint16 `short:"b" long:"buffer-size" default:"8096"`   // Set buffer size
	JSON       bool   `short:"j" long:"json"`                         // Output in JSON format
	Help       bool   `short:"h" long:"help"`                         // Show this help message
	Version    bool   `short:"v" long:"version"`                      // Show version and build info
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
