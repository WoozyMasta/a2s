package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/woozymasta/a2s/internal/vars"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords"
	"github.com/woozymasta/steam/utils/appid"
)

func createClient(host string, timeout int, buffer uint16) *a2s.Client {
	client, err := a2s.NewWithString(host)
	if err != nil {
		fatalf("Failed to create client: %s", err)
	}

	if timeout > 0 {
		client.SetDeadlineTimeout(timeout)
	}
	client.SetBufferSize(buffer)

	return client
}

func printJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}

func printKeywordsJSON(info *a2s.Info) {
	jsonData, err := json.Marshal(info)
	if err != nil {
		fatalf("Failed to marshal Info: %v", err)
	}

	// Unmarshal into a map to add custom fields
	var jsonMap map[string]any
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Add the parsed DayZ structure to the JSON map
	delete(jsonMap, "keywords")

	switch info.ID {
	case appid.Arma3.Uint64():

		armaData := keywords.ParseArma3(info.Keywords)
		jsonMap["keywords"] = armaData

	case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
		dayZData := keywords.ParseDayZ(info.Keywords)
		jsonMap["keywords"] = dayZData
	}

	// Marshal back to JSON for output
	updatedJSONData, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		fatalf("Failed to marshal updated JSON: %v", err)
	}

	fmt.Println(string(updatedJSONData))
}

func printHelp(exit bool) {
	fmt.Printf(`Description:
  CLI for querying A2S server information and working with A3SB subprotocol for Arma 3 and DayZ.

Usage:
  %[1]s [OPTIONS] <command> <host(:query port)> <query port>

Example:
  %[1]s ping 127.0.0.1 27016
  %[1]s -j info 127.0.0.1:27016 | jq '.players'

Commands:
  info     Retrieve server information A2S_INFO;
  rules    Retrieve server rules A2S_RULES;
  players  Retrieve player list A2S_PLAYERS;
  all      Retrieve all available server information;
  ping     Ping the server with A2S_INFO.

Options:
  -j, --json               Output in JSON format;
  -i, --app-id             AppID for more accurate results;
  -t, --deadline-timeout=  Set connection timeout in seconds;
  -b, --buffer-size=       Set connection buffer size;
  -c, --ping-count=        Set the number of ping requests to send;
  -p, --ping-period=       Set the period between pings in seconds;
  -t, --version            Show version, commit, and build time;
  -h, --help               Prints this help message.
`, filepath.Base(os.Args[0]))

	if exit {
		os.Exit(0)
	}
}

func printVersion() {
	fmt.Printf(`file:     %s
version:  %s
commit:   %s
built:    %s
project:  %s
`, os.Args[0], vars.Version, vars.Commit, vars.BuildTime, vars.URL)
	os.Exit(0)
}

func fatal(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
