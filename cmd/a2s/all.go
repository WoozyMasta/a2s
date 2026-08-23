// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"fmt"
)

// executeAll runs the info, rules, and players commands for one server.
func executeAll(app *Application, cmd *AllCommand, clientOptions ClientOptions) error {
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
