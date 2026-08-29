// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/woozymasta/flags"
)

func TestParserHelpContainsConfiguredExamples(t *testing.T) {
	tests := []struct {
		command string
		wants   []string
	}{
		{command: "info", wants: []string{
			"a2s info 127.0.0.1:27015",
			"a2s info 127.0.0.1:27015 --format json | jq '.name, .players, .map'",
		}},
		{command: "players", wants: []string{
			"a2s players 127.0.0.1:27015 --format json | jq '.[] | {name, score}'",
		}},
		{command: "rules", wants: []string{
			"a2s rules example.org:2303 --game arma3",
			"a2s rules example.org:2303 --raw --format json | jq",
		}},
		{command: "all", wants: []string{
			"a2s all 127.0.0.1:27015 --format json",
		}},
		{command: "ping", wants: []string{
			"a2s ping 127.0.0.1:27015 --ping-count 10 --ping-period 2s",
		}},
		{command: "proxy", wants: []string{
			"a2s proxy 127.0.0.1:27015 --listen :27016",
			"a2s proxy 127.0.0.1:27015 --listen :27016 --cache info --cache players --cache rules --ttl 30s",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			parser, _ := newTestParser()
			_, err := parser.ParseArgs([]string{tt.command, "--help"})
			if !isParserError(err, flags.ErrHelp) {
				t.Fatalf("ParseArgs() error = %v, want ErrHelp", err)
			}
			for _, want := range tt.wants {
				if !bytes.Contains([]byte(err.Error()), []byte(want)) {
					t.Fatalf("help does not contain configured example %q:\n%s", want, err)
				}
			}
		})
	}
}

func TestParserEnforcesCommandContract(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want flags.ErrorType
	}{
		{name: "unknown command", args: []string{"infp", "host"}, want: flags.ErrUnknownCommand},
		{name: "unexpected argument", args: []string{"info", "host", "port", "extra"}, want: flags.ErrUnexpectedArgument},
		{name: "missing command", args: nil, want: flags.ErrCommandRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, _ := newTestParser()
			_, err := parser.ParseArgs(tt.args)
			if !isParserError(err, tt.want) {
				t.Fatalf("ParseArgs() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParserAppliesCommonClientOptionsBeforeAndAfterCommand(t *testing.T) {
	tests := [][]string{
		{"--timeout", "750ms", "--buffer-size", "16384", "info", "host"},
		{"info", "host", "--timeout", "750ms", "--buffer-size", "16384"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			parser, options := newTestParser()
			if _, err := parser.ParseArgs(args); err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}

			if time.Duration(options.Timeout) != 750*time.Millisecond {
				t.Fatalf("timeout = %s, want 750ms", options.Timeout)
			}
			if options.Buffer != 16384 {
				t.Fatalf("buffer = %d, want 16384", options.Buffer)
			}
		})
	}
}

func TestParserAcceptsPingOutputOptions(t *testing.T) {
	parser, options := newTestParser()
	if _, err := parser.ParseArgs([]string{
		"ping",
		"host",
		"--query",
		"players",
		"--compact",
		"--no-summary",
	}); err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}

	if !options.Ping.Compact {
		t.Fatal("compact = false, want true")
	}
	if options.Ping.Query != "players" {
		t.Fatalf("query = %q, want players", options.Ping.Query)
	}
	if !options.Ping.NoSummary {
		t.Fatal("no-summary = false, want true")
	}
}

func TestParserAcceptsDurationPingPeriod(t *testing.T) {
	parser, options := newTestParser()
	if _, err := parser.ParseArgs([]string{
		"ping",
		"host",
		"--ping-period",
		"250ms",
	}); err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}

	if time.Duration(options.Ping.PingPeriod) != 250*time.Millisecond {
		t.Fatalf("ping period = %s, want 250ms", options.Ping.PingPeriod)
	}
}

func TestParserTreatsBareDurationAsSeconds(t *testing.T) {
	parser, options := newTestParser()
	if _, err := parser.ParseArgs([]string{
		"ping",
		"host",
		"--ping-period",
		"2",
	}); err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}

	if time.Duration(options.Ping.PingPeriod) != 2*time.Second {
		t.Fatalf("ping period = %s, want 2s", options.Ping.PingPeriod)
	}
}

func TestValidateProxyCommand(t *testing.T) {
	tests := []struct {
		name    string
		command ProxyCommand
		want    []string
		wantErr string
	}{
		{
			name: "default cache",
			command: ProxyCommand{
				Args:   ServerArgs{Host: "127.0.0.1:27015"},
				Listen: ":27016",
			},
			want: []string{"auto"},
		},
		{
			name: "deduplicates concrete values",
			command: ProxyCommand{
				Args:   ServerArgs{Host: "127.0.0.1:27015"},
				Listen: ":27016",
				Cache:  []string{"info", "rules", "info"},
			},
			want: []string{"info", "rules"},
		},
		{
			name: "auto selector cannot be combined",
			command: ProxyCommand{
				Args:   ServerArgs{Host: "127.0.0.1:27015"},
				Listen: ":27016",
				Cache:  []string{"auto", "rules"},
			},
			wantErr: `cache selector "auto" cannot be combined with other values`,
		},
		{
			name: "invalid upstream endpoint",
			command: ProxyCommand{
				Args:   ServerArgs{Host: "127.0.0.1"},
				Listen: ":27016",
			},
			wantErr: "invalid upstream endpoint",
		},
		{
			name: "invalid listen port",
			command: ProxyCommand{
				Args:   ServerArgs{Host: "127.0.0.1:27015"},
				Listen: ":invalid",
			},
			wantErr: "invalid listen port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProxyCommand(&tt.command)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("validateProxyCommand() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateProxyCommand() error = %v", err)
			}
			if strings.Join(tt.command.Cache, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("cache = %v, want %v", tt.command.Cache, tt.want)
			}
		})
	}
}

func newTestParser() (*flags.Parser, *Options) {
	options := &Options{}
	i18nConfig, err := newI18nConfig()
	if err != nil {
		panic(err)
	}
	i18nConfig.Locale = "en"
	parser, err := newParser(options, i18nConfig)
	if err != nil {
		panic(err)
	}
	parser.Name = "a2s"

	return parser, options
}

func isParserError(err error, want flags.ErrorType) bool {
	parserErr, ok := err.(*flags.Error)
	return ok && parserErr.Type == want
}
