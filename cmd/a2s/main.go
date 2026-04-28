// Package main implements the a2s command-line client.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/a2s/internal/vars"
	"github.com/woozymasta/a2s/pkg/a2s"
)

// Options defines the root command structure.
type Options struct {
	Info    InfoCommand    `description:"Retrieve server information A2S_INFO"      command:"info"`
	Players PlayersCommand `description:"Retrieve player list A2S_PLAYERS"          command:"players"`
	Rules   RulesCommand   `description:"Retrieve server rules A2S_RULES"           command:"rules"`
	All     AllCommand     `description:"Retrieve all available server information" command:"all"`
	Ping    PingCommand    `description:"Ping the server with A2S_INFO"             command:"ping"`
	Version bool           `description:"Show version, commit, and build time" short:"v" long:"version"`
}

// InfoCommand handles the 'info' subcommand.
type InfoCommand struct {
	Args ServerArgs `positional-args:"yes"`
	GlobalOptions
}

// PlayersCommand handles the 'players' subcommand.
type PlayersCommand struct {
	Args ServerArgs `positional-args:"yes"`
	GlobalOptions
}

// RulesCommand handles the 'rules' subcommand.
type RulesCommand struct {
	Args ServerArgs `positional-args:"yes"`
	RulesOptions
	GlobalOptions
}

// AllCommand handles the 'all' subcommand.
type AllCommand struct {
	Args ServerArgs `positional-args:"yes"`
	RulesOptions
	GlobalOptions
}

// PingCommand handles the 'ping' subcommand.
type PingCommand struct {
	Args ServerArgs `positional-args:"yes"`
	GlobalOptions
	PingCount  int `short:"c" default:"0" long:"ping-count"  description:"Set the number of ping requests to send (0 = infinite)"`
	PingPeriod int `short:"p" default:"1" long:"ping-period" description:"Set the period between pings in seconds"`
}

// GlobalOptions defines global CLI options applicable to all commands.
type GlobalOptions struct {
	Format  string `short:"f" default:"table" long:"format"      description:"Output format" choice:"json" choice:"table" choice:"raw" choice:"md" choice:"html"`
	Timeout int    `short:"t" default:"3"     long:"timeout"     description:"Set connection timeout in seconds"`
	Buffer  uint16 `short:"b" default:"8096"  long:"buffer-size" description:"Set connection buffer size"`
}

// ServerArgs defines positional arguments for server connection.
type ServerArgs struct {
	Host string `positional-arg-name:"host" description:"Server host (with optional port, e.g., 127.0.0.1:27016)"`
	Port string `positional-arg-name:"port" description:"Query port (if not included in host)"`
}

// RulesOptions defines options specific to rules command.
type RulesOptions struct {
	Game     string `short:"g" long:"game"      description:"Game type for more accurate results" choice:"dayz" choice:"arma3"`
	Raw      bool   `short:"r" long:"raw"       description:"Disable parse A2S_RULES values to types"`
	SkipInfo bool   `short:"s" long:"skip-info" description:"Skip automatic AppID detection via A2S_INFO"`
}

// main parses command-line options and dispatches the selected subcommand.
func main() {
	opts := &Options{}
	p := flags.NewParser(opts, flags.Default)
	p.LongDescription = "CLI for querying Steam A2S server information and working with A3SB subprotocol for Arma 3 and DayZ."
	p.Name = filepath.Base(os.Args[0])

	_, err := p.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if opts.Version {
		vars.Print()
		return
	}

	if p.Active == nil {
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	// Execute the appropriate command
	switch p.Active.Name {
	case "info":
		executeInfo(&opts.Info)
	case "players":
		executePlayers(&opts.Players)
	case "rules":
		executeRules(&opts.Rules)
	case "all":
		executeAll(&opts.All)
	case "ping":
		executePing(&opts.Ping)
	default:
		fatalf("Unknown command: %s", p.Active.Name)
	}
}

// createClient builds and configures a client from CLI connection options.
func createClient(host, port string, timeout int, buffer uint16) *a2s.Client {
	address := host
	if port != "" {
		address = host + ":" + port
	}

	client, err := a2s.NewWithString(address)
	if err != nil {
		fatalf("Failed to create client: %s", err)
	}

	if timeout > 0 {
		client.SetDeadlineTimeout(timeout)
	}
	client.SetBufferSize(buffer)

	return client
}

// closeClient safely closes the client and logs any error.
func closeClient(client *a2s.Client) {
	if err := client.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to close client: %s\n", err)
	}
}

// fatal writes an error message and terminates the CLI with a failure status.
func fatal(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

// fatalf formats an error message and terminates the CLI with a failure status.
func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
