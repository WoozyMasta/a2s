package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/a2s/internal/ping"
)

// CLI options
type Options struct {
	JSON       bool `short:"j" long:"json"`                         // Output in JSON format
	Raw        bool `short:"r" long:"raw"`                          // Disable parse A2S_RULES values to types
	Timeout    int  `short:"t" long:"deadline-timeout" default:"3"` // Set timeout in seconds
	Buffer     int  `short:"b" long:"buffer-size" default:"8096"`   // Set buffer size
	PingCount  int  `short:"c" long:"ping-count" default:"0"`       // Set the number of ping requests to send
	PingPeriod int  `short:"p" long:"ping-period" default:"1"`      // Set the period between pings in seconds
	Help       bool `short:"h" long:"help"`                         // Show this help message
	Version    bool `short:"v" long:"version"`                      // Show version and build info
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
	if len(args) < 3 {
		fatal("Host and port must be provided as positional arguments")
	}
	if len(args) > 3 {
		fatalf("Extra command passed %s", args[3:])
	}

	client := createClient(args[1], args[2], opts.Timeout, opts.Buffer)

	switch args[0] {
	case "info":
		printInfo(client, opts.JSON)
	case "rules":
		printRules(client, opts.JSON, opts.Raw)
	case "players":
		printPlayers(client, opts.JSON)
	case "all":
		printInfo(client, opts.JSON)
		printRules(client, opts.JSON, opts.Raw)
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
