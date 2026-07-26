package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/woozymasta/flags"
)

func TestParserRecognizesCommands(t *testing.T) {
	tests := []string{"info", "players", "rules", "all", "ping"}

	for _, command := range tests {
		t.Run(command, func(t *testing.T) {
			parser, options := newTestParser()
			if _, err := parser.ParseArgs([]string{command, "127.0.0.1"}); err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if parser.Active == nil || parser.Active.Name != command {
				t.Fatalf("active command = %#v, want %q", parser.Active, command)
			}
			if options == nil {
				t.Fatal("parser options are nil")
			}
		})
	}
}

func TestParserRecognizesPrimaryOptions(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(*testing.T, *Options)
	}{
		{
			name: "format",
			args: []string{"info", "--format", "json", "host"},
			check: func(t *testing.T, options *Options) {
				if options.Info.Format != "json" {
					t.Fatalf("format = %q, want json", options.Info.Format)
				}
			},
		},
		{
			name: "rules options",
			args: []string{"rules", "--game", "dayz", "--raw", "host"},
			check: func(t *testing.T, options *Options) {
				if options.Rules.Game != "dayz" || !options.Rules.Raw {
					t.Fatalf("rules options = %#v, want game=dayz raw=true", options.Rules)
				}
			},
		},
		{
			name: "ping options",
			args: []string{"ping", "-c", "5", "-p", "2", "host"},
			check: func(t *testing.T, options *Options) {
				if options.Ping.PingCount != 5 || options.Ping.PingPeriod != 2 {
					t.Fatalf("ping options = %#v, want count=5 period=2", options.Ping)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, options := newTestParser()
			if _, err := parser.ParseArgs(tt.args); err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			tt.check(t, options)
		})
	}
}

func TestParserHelpAndVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "long help", args: []string{"--help"}},
		{name: "short help", args: []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, _ := newTestParser()
			if _, err := parser.ParseArgs(tt.args); !isParserError(err, flags.ErrHelp) {
				t.Fatalf("ParseArgs() error = %v, want ErrHelp", err)
			}
		})
	}

	for _, arg := range []string{"--version", "-v"} {
		t.Run(arg, func(t *testing.T) {
			parser, options := newTestParser()
			if _, err := parser.ParseArgs([]string{arg}); !isParserError(err, flags.ErrVersion) {
				t.Fatalf("ParseArgs() error = %v, want ErrVersion", err)
			}
			if options.Info.Args.Host != "" {
				t.Fatalf("version parsing unexpectedly populated command arguments for %s", arg)
			}
		})
	}
}

func TestParserAcceptsChoices(t *testing.T) {
	for _, format := range []string{"json", "table", "raw", "md", "html"} {
		t.Run("format/"+format, func(t *testing.T) {
			parser, options := newTestParser()
			if _, err := parser.ParseArgs([]string{"info", "--format", format, "host"}); err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if options.Info.Format != format {
				t.Fatalf("format = %q, want %q", options.Info.Format, format)
			}
		})
	}

	for _, game := range []string{"dayz", "arma3"} {
		t.Run("game/"+game, func(t *testing.T) {
			parser, options := newTestParser()
			if _, err := parser.ParseArgs([]string{"rules", "--game", game, "host"}); err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if options.Rules.Game != game {
				t.Fatalf("game = %q, want %q", options.Rules.Game, game)
			}
		})
	}
}

func TestParserRejectsInvalidChoices(t *testing.T) {
	tests := [][]string{
		{"info", "--format", "yaml", "host"},
		{"rules", "--game", "rust", "host"},
	}

	for _, args := range tests {
		t.Run(args[1], func(t *testing.T) {
			parser, _ := newTestParser()
			if _, err := parser.ParseArgs(args); !isParserError(err, flags.ErrInvalidChoice) {
				t.Fatalf("ParseArgs() error = %v, want ErrInvalidChoice", err)
			}
		})
	}
}

func TestParserStrictValidation(t *testing.T) {
	parser, _ := newTestParser()
	_, err := parser.ParseArgs([]string{"infp", "host"})
	if !isParserError(err, flags.ErrUnknownCommand) {
		t.Fatalf("unknown command error = %v, want ErrUnknownCommand", err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "info") {
		t.Fatalf("unknown command error = %v, want suggestion for info", err)
	}

	parser, _ = newTestParser()
	_, err = parser.ParseArgs([]string{"info", "host", "port", "extra"})
	if !isParserError(err, flags.ErrUnexpectedArgument) {
		t.Fatalf("extra argument error = %v, want ErrUnexpectedArgument", err)
	}

	parser, options := newTestParser()
	if _, err := parser.ParseArgs([]string{"info", "--", "host"}); err != nil {
		t.Fatalf("double-dash parse error = %v", err)
	}
	if options.Info.Args.Host != "host" {
		t.Fatalf("host = %q, want host after --", options.Info.Args.Host)
	}
}

func TestParserValidatesNumericOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "negative count", args: []string{"ping", "--ping-count", "-1", "host"}},
		{name: "zero period", args: []string{"ping", "--ping-period", "0", "host"}},
		{name: "negative period", args: []string{"ping", "--ping-period", "-1", "host"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, _ := newTestParser()
			if _, err := parser.ParseArgs(tt.args); !isParserError(err, flags.ErrValidation) {
				t.Fatalf("ParseArgs() error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestParserRequiresHostForNetworkCommands(t *testing.T) {
	for _, command := range []string{"info", "players", "rules", "all", "ping"} {
		t.Run(command, func(t *testing.T) {
			parser, _ := newTestParser()
			if _, err := parser.ParseArgs([]string{command}); !isParserError(err, flags.ErrRequired) {
				t.Fatalf("ParseArgs() error = %v, want ErrRequired", err)
			}
		})
	}
}

func TestParserRootWithoutCommand(t *testing.T) {
	parser, _ := newTestParser()
	if _, err := parser.ParseArgs(nil); !isParserError(err, flags.ErrCommandRequired) {
		t.Fatalf("ParseArgs() error = %v, want ErrCommandRequired", err)
	}
	if parser.Active != nil {
		t.Fatalf("active command = %#v, want nil", parser.Active)
	}
}

func TestParserHelpContainsPrimaryCommandsAndOptions(t *testing.T) {
	parser, _ := newTestParser()
	var help bytes.Buffer
	parser.WriteHelp(&help)

	for _, want := range []string{"info", "players", "rules", "all", "ping", "version"} {
		if !bytes.Contains(help.Bytes(), []byte(want)) {
			t.Fatalf("help does not contain %q:\n%s", want, help.String())
		}
	}

	parser, _ = newTestParser()
	_, err := parser.ParseArgs([]string{"info", "--help"})
	if !isParserError(err, flags.ErrHelp) {
		t.Fatalf("command help error = %v, want ErrHelp", err)
	}
	if !bytes.Contains([]byte(err.Error()), []byte("format")) {
		t.Fatalf("command help does not contain format option: %v", err)
	}
}

func TestParserCommandsHaveLongDescriptions(t *testing.T) {
	want := map[string]string{
		"info":    "Query server metadata with A2S_INFO.",
		"players": "Query the current player list with A2S_PLAYER.",
		"rules":   "Query server rules with A2S_RULES or automatic A3SB parsing.",
		"all":     "Query server metadata, rules, and players in one command.",
		"ping":    "Measure server response time with repeated A2S_INFO queries.",
	}

	parser, _ := newTestParser()
	for name, description := range want {
		t.Run(name, func(t *testing.T) {
			command := parser.Command.Find(name)
			if command == nil {
				t.Fatalf("command %q was not found", name)
			}
			if command.LongDescription != description {
				t.Fatalf("long description = %q, want %q", command.LongDescription, description)
			}
		})
	}
}

func TestParserHelpContainsStructuredExamples(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{command: "info", want: "a2s info 127.0.0.1:27015"},
		{command: "rules", want: "a2s rules example.org:2303 --game arma3"},
		{command: "all", want: "a2s all 127.0.0.1:27015 --format json"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			parser, _ := newTestParser()
			_, err := parser.ParseArgs([]string{tt.command, "--help"})
			if !isParserError(err, flags.ErrHelp) {
				t.Fatalf("ParseArgs() error = %v, want ErrHelp", err)
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tt.want)) {
				t.Fatalf("help does not contain structured example %q:\n%s", tt.want, err)
			}
		})
	}
}

func newTestParser() (*flags.Parser, *Options) {
	options := &Options{}
	parser, err := newParser(options)
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
