// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

// Package main implements the a2s command-line client.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/flags"
)

// Options defines the root command structure.
type Options struct {
	Rules   RulesCommand   `command:"rules"   ini-group:"rules"   command-i18n:"command.rules.description"`
	All     AllCommand     `command:"all"     ini-group:"all"     command-i18n:"command.all.description"`
	Info    InfoCommand    `command:"info"    ini-group:"info"    command-i18n:"command.info.description"`
	Players PlayersCommand `command:"players" ini-group:"players" command-i18n:"command.players.description"`
	Ping    PingCommand    `command:"ping"    ini-group:"ping"    command-i18n:"command.ping.description"`
	Proxy   ProxyCommand   `command:"proxy"   ini-group:"proxy"   command-i18n:"command.proxy.description"`
	ClientOptions
}

// InfoCommand handles the 'info' subcommand.
type InfoCommand struct {
	Args ServerArgs `positional-args:"yes"`
	OutputOptions
}

// PlayersCommand handles the 'players' subcommand.
type PlayersCommand struct {
	Args ServerArgs `positional-args:"yes"`
	OutputOptions
}

// RulesCommand handles the 'rules' subcommand.
type RulesCommand struct {
	Args ServerArgs `positional-args:"yes"`
	OutputOptions
	RulesOptions
}

// AllCommand handles the 'all' subcommand.
type AllCommand struct {
	Args ServerArgs `positional-args:"yes"`
	OutputOptions
	RulesOptions
}

// PingCommand handles the 'ping' subcommand.
type PingCommand struct {
	Args ServerArgs `positional-args:"yes"`

	PingCount  int `short:"c" default:"0" validate-min:"0" long:"ping-count"  description-i18n:"option.ping_count.description"`
	PingPeriod int `short:"p" default:"1" validate-min:"1" long:"ping-period" description-i18n:"option.ping_period.description"`
}

// ProxyCommand defines configuration for the cached A2S proxy.
type ProxyCommand struct {
	Args ServerArgs `positional-args:"yes"`

	Listen       string        `long:"listen"        description-i18n:"option.proxy.listen.description" required:"true"`
	Cache        []string      `long:"cache"         description-i18n:"option.proxy.cache.description"         default:"auto" choices:"info;players;rules;auto"`
	TTL          time.Duration `long:"ttl"           description-i18n:"option.proxy.ttl.description"           default:"15s"  validate-min:"1"`
	InactiveTTL  time.Duration `long:"inactive-ttl"  description-i18n:"option.proxy.inactive_ttl.description"  default:"0"    validate-min:"0"`
	Jitter       time.Duration `long:"jitter"        description-i18n:"option.proxy.jitter.description"        default:"1s"   validate-min:"0"`
	Retries      int           `long:"retries"       description-i18n:"option.proxy.retries.description"       default:"2"    validate-min:"0"`
	UpstreamPing bool          `long:"upstream-ping" description-i18n:"option.proxy.upstream_ping.description"`
}

// ClientOptions defines network settings shared by all network commands.
type ClientOptions struct {
	Timeout time.Duration `short:"t" default:"3s"   validate-min:"1" long:"timeout"     description-i18n:"option.timeout.description"`
	Buffer  uint16        `short:"b" default:"8192" validate-min:"1" long:"buffer-size" description-i18n:"option.buffer_size.description"`
}

// OutputOptions defines output settings for commands that render responses.
type OutputOptions struct {
	Format string `short:"f" long:"format" default:"table" choices:"json;table;raw;md;html" description-i18n:"option.format.description"`
}

// ServerArgs defines positional arguments for server connection.
type ServerArgs struct {
	Host string `positional-arg-name:"host" arg-name-i18n:"arg.host.name" arg-description-i18n:"arg.host.description" required:"true"`
	Port string `positional-arg-name:"port" arg-name-i18n:"arg.port.name" arg-description-i18n:"arg.port.description"`
}

// RulesOptions defines options specific to rules command.
type RulesOptions struct {
	Game string `short:"g" long:"game" description-i18n:"option.game.description" choices:"dayz;arma3"`
	Raw  bool   `short:"r" long:"raw"  description-i18n:"option.raw.description"`
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
	i18nConfig, err := newI18nConfig()
	if err != nil {
		return err
	}
	if app != nil {
		app.Localizer = flags.NewLocalizer(i18nConfig)
	}

	p, err := newParser(opts, i18nConfig)
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
		return fmt.Errorf("%s", app.localize("error.no_command", "no command specified"))
	}

	// Execute the appropriate command
	switch p.Active.Name {
	case "help", "version", "completion", "docs":
		return nil // Built-in commands execute during ParseArgs.

	case "info":
		return executeInfo(app, &opts.Info, opts.ClientOptions)

	case "players":
		return executePlayers(app, &opts.Players, opts.ClientOptions)

	case "rules":
		return executeRules(app, &opts.Rules, opts.ClientOptions)

	case "all":
		return executeAll(app, &opts.All, opts.ClientOptions)

	case "ping":
		return executePing(app, &opts.Ping, opts.ClientOptions)

	case "proxy":
		return executeProxy(app, &opts.Proxy, opts.ClientOptions)

	default:
		return fmt.Errorf("%s", app.localize(
			"error.unknown_command",
			"unknown command: %s",
			p.Active.Name,
		))
	}
}

// newParser creates the CLI parser with a shared localization config.
func newParser(opts *Options, i18nConfig flags.I18nConfig) (*flags.Parser, error) {
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
	parser.SetI18n(i18nConfig)

	parser.SetVersionInfo(flags.VersionInfo{
		Version:      version,
		Revision:     commit,
		RevisionTime: buildTime,
		URL:          repositoryURL,
	})
	parser.SetVersionFields(flags.VersionFieldsCore)

	parser.SetLongDescriptionI18nKey("cli.description")
	parser.SetCommandLongDescriptionI18nKeys(map[string]string{
		"info":    "command.info.long",
		"players": "command.players.long",
		"rules":   "command.rules.long",
		"all":     "command.all.long",
		"ping":    "command.ping.long",
		"proxy":   "command.proxy.long",
	})

	return parser, parser.SetCommandExamples(map[string][]*flags.CommandExample{
		"info": {
			flags.Example().
				Arg("127.0.0.1:27015"),
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.Info.Format, "json").
				Raw("| jq '.name, .players, .map'"),
		},

		"players": {
			flags.Example().
				Arg("127.0.0.1:27015"),
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.Players.Format, "json").
				Raw("| jq '.[] | {name, score}'"),
		},

		"rules": {
			flags.Example().
				Arg("example.org:2303").
				Option(&opts.Rules.Game, "arma3"),
			flags.Example().
				Arg("example.org:2303").
				Option(&opts.Rules.Raw).
				Option(&opts.Rules.Format, "json").
				Raw("| jq 'to_entries[] | \"\\(.key)=\\(.value)\"'"),
		},

		"all": {
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.All.Format, "json"),
		},

		"ping": {
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.Ping.PingCount, "5"),
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.Ping.PingCount, "10").
				Option(&opts.Ping.PingPeriod, "2"),
		},

		"proxy": {
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.Proxy.Listen, ":27016"),
			flags.Example().
				Arg("127.0.0.1:27015").
				Option(&opts.Proxy.Listen, ":27016").
				Option(&opts.Proxy.Cache, "info").
				Option(&opts.Proxy.Cache, "players").
				Option(&opts.Proxy.Cache, "rules").
				Option(&opts.Proxy.TTL, "30s"),
		},
	})
}

// createClient builds and configures a client from CLI connection options.
func createClient(host, port string, timeout time.Duration, buffer uint16) (*a2s.Client, error) {
	address, err := normalizeEndpoint(host, port)
	if err != nil {
		return nil, err
	}

	options := []a2s.Option{a2s.WithBufferSize(buffer)}
	if timeout > 0 {
		options = append(options, a2s.WithTimeout(timeout))
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
		_, _ = fmt.Fprintf(
			app.Err,
			"%s: %s\n",
			app.localize("warning.close_client", "Warning: failed to close client"),
			err,
		)
	}
}
