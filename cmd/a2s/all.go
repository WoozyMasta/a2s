// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"fmt"
)

// executeAll runs the info, rules, and players commands for one server.
func executeAll(app *Application, cmd *AllCommand, clientOptions ClientOptions) error {
	formatter := NewFormatter(cmd.Format, app.Out, app.Localizer)
	if formatter.ShouldUseJSON() {
		return executeAllJSON(app, cmd, clientOptions, formatter)
	}

	// Execute info
	infoCmd := InfoCommand{
		OutputOptions: cmd.OutputOptions,
	}
	infoCmd.Args.Host = cmd.Args.Host
	infoCmd.Args.Port = cmd.Args.Port
	if err := executeInfo(app, &infoCmd, clientOptions); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(app.Out)

	// Execute rules
	rulesCmd := RulesCommand{
		OutputOptions: cmd.OutputOptions,
		RulesOptions:  cmd.RulesOptions,
	}
	rulesCmd.Args.Host = cmd.Args.Host
	rulesCmd.Args.Port = cmd.Args.Port
	if err := executeRules(app, &rulesCmd, clientOptions); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(app.Out)

	// Execute players
	playersCmd := PlayersCommand{
		OutputOptions: cmd.OutputOptions,
	}
	playersCmd.Args.Host = cmd.Args.Host
	playersCmd.Args.Port = cmd.Args.Port

	return executePlayers(app, &playersCmd, clientOptions)
}

// executeAllJSON queries all aggregate values and emits one JSON document.
func executeAllJSON(
	app *Application,
	cmd *AllCommand,
	clientOptions ClientOptions,
	formatter *Formatter,
) error {
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

	ctx := context.Background()
	info, _, err := client.GetInfoWithMeta(ctx)
	if err != nil {
		return friendlyQueryError(app, "error.server_info", "failed to get server info", err, client.Timeout())
	}
	infoJSON, err := infoJSONValue(info, formatter)
	if err != nil {
		return err
	}

	rules, err := queryRulesValue(ctx, app, client, &RulesCommand{
		RulesOptions: cmd.RulesOptions,
	})
	if err != nil {
		return err
	}

	players, err := client.GetPlayers(ctx)
	if err != nil {
		return friendlyQueryError(app, "error.players", "failed to get players", err, client.Timeout())
	}

	return formatter.PrintJSON(map[string]any{
		"info":    infoJSON,
		"rules":   rules,
		"players": players,
	})
}
