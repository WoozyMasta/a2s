package main

import (
	"github.com/woozymasta/a2s/internal/ping"
)

func executePing(cmd *PingCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer closeClient(client)

	ping.Start(client, cmd.PingCount, cmd.PingPeriod)
}
