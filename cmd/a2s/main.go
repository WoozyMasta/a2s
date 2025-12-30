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
	Help    bool `short:"h" long:"help" description:"Show this help message"`
	Version bool `short:"v" long:"version" description:"Show version, commit, and build time"`

	Info    InfoCommand    `command:"info" description:"Retrieve server information A2S_INFO"`
	Players PlayersCommand `command:"players" description:"Retrieve player list A2S_PLAYERS"`
	Rules   RulesCommand   `command:"rules" description:"Retrieve server rules A2S_RULES"`
	All     AllCommand     `command:"all" description:"Retrieve all available server information"`
	Ping    PingCommand    `command:"ping" description:"Ping the server with A2S_INFO"`
}

// InfoCommand handles the 'info' subcommand.
type InfoCommand struct {
	GlobalOptions

	Args ServerArgs `positional-args:"yes"`
}

// PlayersCommand handles the 'players' subcommand.
type PlayersCommand struct {
	GlobalOptions

	Args ServerArgs `positional-args:"yes"`
}

// RulesCommand handles the 'rules' subcommand.
type RulesCommand struct {
	GlobalOptions
	RulesOptions

	Args ServerArgs `positional-args:"yes"`
}

// AllCommand handles the 'all' subcommand.
type AllCommand struct {
	GlobalOptions
	RulesOptions

	Args ServerArgs `positional-args:"yes"`
}

// PingCommand handles the 'ping' subcommand.
type PingCommand struct {
	GlobalOptions

	PingCount  int        `short:"c" long:"ping-count" default:"0" description:"Set the number of ping requests to send (0 = infinite)"`
	PingPeriod int        `short:"p" long:"ping-period" default:"1" description:"Set the period between pings in seconds"`
	Args       ServerArgs `positional-args:"yes"`
}

// GlobalOptions defines global CLI options applicable to all commands.
type GlobalOptions struct {
	Timeout int    `short:"t" long:"timeout" default:"3" description:"Set connection timeout in seconds"`
	Buffer  uint16 `short:"b" long:"buffer-size" default:"8096" description:"Set connection buffer size"`
	Format  string `short:"f" long:"format" default:"table" description:"Output format" choice:"json" choice:"table" choice:"raw" choice:"md" choice:"html"`
}

// ServerArgs defines positional arguments for server connection.
type ServerArgs struct {
	Host string `positional-arg-name:"host" description:"Server host (with optional port, e.g., 127.0.0.1:27016)"`
	Port string `positional-arg-name:"port" description:"Query port (if not included in host)"`
}

// RulesOptions defines options specific to rules command.
type RulesOptions struct {
	Raw      bool   `short:"r" long:"raw" description:"Disable parse A2S_RULES values to types"`
	Game     string `short:"g" long:"game" description:"Game type for more accurate results" choice:"dayz" choice:"arma3"`
	SkipInfo bool   `short:"s" long:"skip-info" description:"Skip automatic AppID detection via A2S_INFO"`
}

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

	// Handle root-level help/version flags
	if opts.Help {
		p.WriteHelp(os.Stdout)
		return
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

func fatal(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
