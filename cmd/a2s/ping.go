package main

import (
	"github.com/woozymasta/a2s/internal/ping"
)

// executePing runs the configured ping loop for one server.
func executePing(cmd *PingCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer closeClient(client)

	ping.Start(client, cmd.PingCount, cmd.PingPeriod)
}
