// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import "github.com/woozymasta/a2s/internal/ping"

// executePing runs the configured ping loop for one server.
func executePing(app *Application, cmd *PingCommand, clientOptions ClientOptions) error {
	client, err := createClient(
		cmd.Args.Host,
		cmd.Args.Port,
		clientOptions.Timeout,
		clientOptions.Buffer,
	)
	if err != nil {
		return app.wrapError("error.client_create", "failed to create client", err)
	}
	defer closeClient(app, client)

	ping.Start(client, cmd.PingCount, cmd.PingPeriod, ping.Output{
		Out:      app.Out,
		Err:      app.Err,
		Localize: app.localize,
	})

	return nil
}
