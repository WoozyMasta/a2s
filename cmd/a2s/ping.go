package main

import "github.com/woozymasta/a2s/internal/ping"

// executePing runs the configured ping loop for one server.
func executePing(app *Application, cmd *PingCommand) error {
	client, err := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
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
