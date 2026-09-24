// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"time"

	"github.com/woozymasta/a2s/internal/ping"
	"github.com/woozymasta/a2s/pkg/a2s"
)

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

	ping.Start(client, cmd.PingCount, time.Duration(cmd.PingPeriod), pingQueryType(cmd.Query), ping.Output{
		Out:      app.Out,
		Err:      app.Err,
		Localize: app.localize,
		LocalizeLabel: func(key, fallback string) string {
			if app != nil && app.Localizer != nil {
				return app.Localizer.Localize(key, fallback, nil)
			}
			return fallback
		},
		Compact:      cmd.Compact,
		NoSummary:    cmd.NoSummary,
		FormatGameID: formatAppID,
	})

	return nil
}

// pingQueryType converts the CLI query selector to an A2S request type.
func pingQueryType(query string) a2s.QueryType {
	switch query {
	case "players":
		return a2s.PlayerRequest

	case "rules":
		return a2s.RulesRequest

	default:
		return a2s.InfoRequest
	}
}
