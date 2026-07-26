// Package main implements the a2s command-line client.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/woozymasta/a2s/internal/vars"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/flags"
)

// Options defines the root command structure.
type Options struct {
	Info    InfoCommand    `command:"info"    description:"Retrieve server information A2S_INFO"`
	Players PlayersCommand `command:"players" description:"Retrieve player list A2S_PLAYERS"`
	Rules   RulesCommand   `command:"rules"   description:"Retrieve server rules A2S_RULES"`
	All     AllCommand     `command:"all"     description:"Retrieve all available server information"`
	Ping    PingCommand    `command:"ping"    description:"Ping the server with A2S_INFO"`
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
	PingCount  int `short:"c" default:"0" validate-min:"0" long:"ping-count"  description:"Set the number of ping requests to send (0 = infinite)"`
	PingPeriod int `short:"p" default:"1" validate-min:"1" long:"ping-period" description:"Set the period between pings in seconds"`
}

// GlobalOptions defines global CLI options applicable to all commands.
type GlobalOptions struct {
	Format  string `short:"f" default:"table" long:"format"      description:"Output format" choices:"json;table;raw;md;html"`
	Timeout int    `short:"t" default:"3"     long:"timeout"     description:"Set connection timeout in seconds"`
	Buffer  uint16 `short:"b" default:"8192"  long:"buffer-size" description:"Set connection buffer size"`
}

// ServerArgs defines positional arguments for server connection.
type ServerArgs struct {
	Host string `positional-arg-name:"host" description:"Server host (with optional port, e.g., 127.0.0.1:27016)" required:"true"`
	Port string `positional-arg-name:"port" description:"Query port (if not included in host)"`
}

// RulesOptions defines options specific to rules command.
type RulesOptions struct {
	Game string `short:"g" long:"game" description:"Game type for more accurate results" choices:"dayz;arma3"`
	Raw  bool   `short:"r" long:"raw"  description:"Disable parse A2S_RULES values to types"`
}

// main parses command-line options and dispatches the selected subcommand.
func main() {
	app := NewApplication(os.Stdout, os.Stderr)
	if err := run(os.Args[1:], app); err != nil {
		if _, ok := err.(*flags.Error); !ok {
			_, _ = fmt.Fprintln(app.Err, err)
		}
		os.Exit(1)
	}
}

// run parses command-line options and executes the selected subcommand.
func run(args []string, app *Application) error {
	opts := &Options{}
	p, err := newParser(opts)
	if err != nil {
		return err
	}

	p.Name = filepath.Base(os.Args[0])

	_, err = p.ParseArgs(args)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok &&
			(flagsErr.Type == flags.ErrHelp || flagsErr.Type == flags.ErrVersion) {
			return nil
		}
		return err
	}

	if p.Active == nil {
		p.WriteHelp(app.Out)
		return fmt.Errorf("no command specified")
	}

	// Execute the appropriate command
	switch p.Active.Name {
	case "help", "version", "completion":
		// Built-in flags commands execute during ParseArgs.
		return nil
	case "info":
		return executeInfo(app, &opts.Info)
	case "players":
		return executePlayers(app, &opts.Players)
	case "rules":
		return executeRules(app, &opts.Rules)
	case "all":
		return executeAll(app, &opts.All)
	case "ping":
		return executePing(app, &opts.Ping)
	default:
		return fmt.Errorf("unknown command: %s", p.Active.Name)
	}
}

// newParser creates the CLI parser and configures its built-in commands.
func newParser(opts *Options) (*flags.Parser, error) {
	parser := flags.NewParser(opts,
		flags.Default|
			flags.DetectShellFlagStyle|
			flags.DetectShellEnvStyle|
			flags.PrintHelpOnInputErrors|
			flags.HelpCommand|
			flags.VersionCommand|
			flags.CompletionCommand|
			flags.DocsCommand|
			flags.VersionFlag|
			flags.StrictPositionalArgs|
			flags.ShowRepeatableInHelp,
	)

	parser.LongDescription = "CLI for querying Steam A2S server information " +
		"and working with A3SB subprotocol for Arma 3 and DayZ."

	parser.SetVersionInfo(flags.VersionInfo{
		Version:      vars.Version,
		Revision:     vars.Commit,
		RevisionTime: vars.BuildTime,
		URL:          vars.URL,
	})
	parser.SetVersionFields(flags.VersionFieldsCore)
	parser.SetCommandLongDescriptions(map[string]string{
		"info":    "Query server metadata with A2S_INFO.",
		"players": "Query the current player list with A2S_PLAYER.",
		"rules":   "Query server rules with A2S_RULES or automatic A3SB parsing.",
		"all":     "Query server metadata, rules, and players in one command.",
		"ping":    "Measure server response time with repeated A2S_INFO queries.",
	})

	return parser, parser.SetCommandExamples(map[string][]*flags.CommandExample{
		"info":    {flags.Example().Arg("127.0.0.1:27015")},
		"players": {flags.Example().Arg("127.0.0.1:27015")},
		"rules":   {flags.Example().Arg("example.org:2303").Option(&opts.Rules.Game, "arma3")},
		"all":     {flags.Example().Arg("127.0.0.1:27015").Option(&opts.All.Format, "json")},
		"ping":    {flags.Example().Arg("127.0.0.1:27015").Option(&opts.Ping.PingCount, "5")},
	})
}

// createClient builds and configures a client from CLI connection options.
func createClient(host, port string, timeout int, buffer uint16) (*a2s.Client, error) {
	address, err := normalizeEndpoint(host, port)
	if err != nil {
		return nil, err
	}

	options := []a2s.Option{a2s.WithBufferSize(buffer)}
	if timeout > 0 {
		options = append(options, a2s.WithTimeout(time.Duration(timeout)*time.Second))
	}

	client, err := a2s.NewWithString(address, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// closeClient safely closes the client and logs any error.
func closeClient(app *Application, client *a2s.Client) {
	if err := client.Close(); err != nil {
		_, _ = fmt.Fprintf(app.Err, "Warning: failed to close client: %s\n", err)
	}
}
