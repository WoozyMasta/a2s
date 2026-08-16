// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
	"github.com/woozymasta/a2s/pkg/appid"
)

// gameToAppID converts a game name string to AppID.
// Returns 0 if the game name is not recognized.
func gameToAppID(game string) uint64 {
	switch strings.ToLower(game) {
	case "arma3", "arma":
		return appid.Arma3

	case "dayz":
		return appid.DayZ

	default:
		return 0
	}
}

// executeRules selects the standard or automatic A3SB rules parser and renders its output.
// Automatic mode uses one A2S_RULES request and does not require A2S_INFO merely to choose a parser.
func executeRules(app *Application, cmd *RulesCommand) error {
	client, err := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	if err != nil {
		return app.wrapError("error.client_create", "failed to create client", err)
	}
	defer closeClient(app, client)

	formatter := NewFormatter(cmd.Format, app.Out, app.Localizer)
	ctx := context.Background()

	if cmd.Game != "" {
		appID := gameToAppID(cmd.Game)
		if appID == 0 {
			return fmt.Errorf(
				"%s",
				app.localize(
					"error.unknown_game",
					"unknown game: %s. Supported games: arma3, dayz",
					cmd.Game,
				),
			)
		}
		if cmd.Raw {
			return executeRulesStandard(app, ctx, client, true, formatter)
		}

		return executeRulesA3SB(app, ctx, client, appID, formatter)
	}

	if cmd.Raw {
		return executeRulesStandard(app, ctx, client, cmd.Raw, formatter)
	}

	return executeRulesA3SB(app, ctx, client, 0, formatter)
}

// executeRulesStandard retrieves and renders ordinary A2S_RULES values.
func executeRulesStandard(
	app *Application,
	ctx context.Context,
	client *a2s.Client,
	raw bool,
	formatter *Formatter,
) error {
	var rules map[string]any
	var err error

	if raw {
		rawRules, rawErr := client.GetRules(ctx)
		err = rawErr
		rules = make(map[string]any, len(rawRules))
		for _, rule := range rawRules {
			rules[rule.Name] = rule.Value
		}
	} else {
		rules, err = client.GetParsedRules(ctx)
	}

	if err != nil {
		return friendlyQueryError(app, "error.rules", "failed to get rules", err, client.Timeout())
	}

	return printRules(app, rules, client, formatter)
}

// printRules renders an already fetched rules map using the normal A2S output shape.
// Values remain typed for JSON and are formatted as text in tables.
func printRules(app *Application, rules map[string]any, client *a2s.Client, formatter *Formatter) error {
	if formatter.ShouldUseJSON() {
		return formatter.PrintJSON(rules)
	}

	return renderRulesTable(app, rules, client.Addr().String(), formatter)
}

// renderRulesTable renders ordinary A2S rules without querying a server.
func renderRulesTable(app *Application, rules map[string]any, address string, formatter *Formatter) error {
	header := table.Row{
		app.localize("table.rule", "Rule"),
		app.localize("table.value", "Value"),
	}
	rows := make([]table.Row, 0, len(rules))

	// Sort keys so table and text output is deterministic.
	keys := make([]string, 0, len(rules))
	for k := range rules {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		value := rules[k]
		if boolValue, ok := value.(bool); ok {
			value = app.formatBool(boolValue)
		}
		rows = append(rows, table.Row{k, value})
	}

	t := formatter.NewTable(header, rows)
	if err := formatter.PrintTable(t); err != nil {
		return app.wrapError("error.render_rules", "failed to render rules", err)
	}

	if formatter.IsTableFormat() {
		_, _ = fmt.Fprintf(
			app.Out,
			"%s\n",
			app.localize("footer.rules", "A2S_RULES response for %s", address),
		)
	}

	return nil
}

// executeRulesA3SB retrieves and renders Arma 3/DayZ server-browser rules.
func executeRulesA3SB(
	app *Application,
	ctx context.Context,
	client *a2s.Client,
	appID uint64,
	formatter *Formatter,
) error {
	a3sbClient := &a3sb.Client{Client: client}

	rules, err := a3sbClient.GetRules(ctx, appID)
	if err != nil {
		return friendlyQueryError(app, "error.server_rules", "failed to get server rules", err, client.Timeout())
	}

	if rules.Version == 0 {
		parsed := a2s.ParseRuleValues(rules.ExtraRules)
		return printRules(app, parsed, client, formatter)
	}

	if formatter.ShouldUseJSON() {
		return formatter.PrintJSON(rules)
	}

	return renderA3SBRules(app, rules, formatter, client.Addr().String())
}

// renderA3SBRules renders human-readable A3SB fields without querying a server.
func renderA3SBRules(app *Application, rules *a3sb.Rules, formatter *Formatter, address string) error {
	// Print Island/Description info (DayZ specific)
	if rules.Island != "" {
		formatter.PrintSectionHeader(app.localize("section.server_information", "Server Information"))
		header := table.Row{
			app.localize("table.option", "Option"),
			app.localize("table.value", "Value"),
		}
		rows := make([]table.Row, 0, 10)

		if rules.Description != "" {
			rows = append(rows, table.Row{
				app.localize("rules.description", "Description:"),
				rules.Description,
			})
		}

		rows = append(rows, []table.Row{
			{
				app.localize("rules.allowed_build", "Allowed build:"),
				fmt.Sprintf("%d", rules.AllowedBuild),
			},
			{
				app.localize("rules.client_port", "Client port:"),
				fmt.Sprintf("%d", rules.ClientPort),
			},
			{
				app.localize("rules.dedicated", "Dedicated:"),
				fmt.Sprintf("%t", rules.Dedicated),
			},
			{
				app.localize("rules.island", "Island:"),
				rules.Island,
			},
			{
				app.localize("rules.language", "Language:"),
				rules.Language.String(),
			},
			{
				app.localize("rules.platform", "Platform:"),
				rules.Platform,
			},
			{
				app.localize("rules.required_build", "Required build:"),
				fmt.Sprintf("%d", rules.RequiredBuild),
			},
			{
				app.localize("rules.required_version", "Required version:"),
				fmt.Sprintf("%d", rules.RequiredVersion),
			},
			{
				app.localize("rules.time_left", "TimeLeft:"),
				fmt.Sprintf("%d", rules.TimeLeft),
			},
		}...)

		t := formatter.NewTable(header, rows)
		if err := formatter.PrintTable(t); err != nil {
			return app.wrapError("error.render_rules", "failed to render rules", err)
		}
	}

	// Print Difficulty (Arma3 specific)
	if rules.Difficulty != nil {
		formatter.PrintSectionHeader(app.localize("section.difficulty", "Difficulty Settings"))
		header := table.Row{
			app.localize("table.option", "Option"),
			app.localize("table.value", "Value"),
		}

		rows := []table.Row{
			{
				app.localize("rules.difficulty_level", "Difficulty Level:"),
				fmt.Sprintf("%d", rules.Difficulty.Level),
			},
			{
				app.localize("rules.ai_level", "AI Level:"),
				fmt.Sprintf("%d", rules.Difficulty.AILevel),
			},
			{
				app.localize("rules.advanced_flight", "Advanced Flight:"),
				fmt.Sprintf("%t", rules.Difficulty.AdvanceFlight),
			},
			{
				app.localize("rules.third_person", "Third Person:"),
				fmt.Sprintf("%t", rules.Difficulty.ThirdPerson),
			},
			{
				app.localize("rules.crosshair", "Crosshair:"),
				fmt.Sprintf("%t", rules.Difficulty.Crosshair),
			},
		}

		t := formatter.NewTable(header, rows)
		if err := formatter.PrintTable(t); err != nil {
			return app.wrapError("error.render_rules", "failed to render rules", err)
		}
	}

	// Print DLC
	if len(rules.DLC) > 0 {
		formatter.PrintSectionHeader(app.localize("section.dlc", "DLC"))
		header := table.Row{
			app.localize("table.number", "#"),
			app.localize("rules.dlc_name", "DLC Name"),
			app.localize("rules.dlc_url", "DLC URL"),
		}
		rows := make([]table.Row, 0, len(rules.DLC))

		for i, dlc := range rules.DLC {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				dlc.Name,
				fmt.Sprintf("https://store.steampowered.com/app/%d", dlc.ID),
			})
		}
		if shouldCompactLinkTable(formatter, header, rows) {
			header[2] = app.localize("rules.app_id", "APP ID")
			for index, dlc := range rules.DLC {
				rows[index][2] = fmt.Sprintf("%d", dlc.ID)
			}
		}

		t := formatter.NewTable(header, rows)
		if err := formatter.PrintTable(t); err != nil {
			return app.wrapError("error.render_rules", "failed to render rules", err)
		}
	}

	// Print Creator DLC
	if len(rules.CreatorDLC) > 0 {
		formatter.PrintSectionHeader(app.localize("section.creator_dlc", "Creator DLC"))
		header := table.Row{
			app.localize("table.number", "#"),
			app.localize("rules.creator_dlc_name", "Creator DLC Name"),
			app.localize("rules.creator_dlc_url", "Creator DLC URL"),
		}
		rows := make([]table.Row, 0, len(rules.CreatorDLC))

		for i, dlc := range rules.CreatorDLC {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				dlc.Name,
				fmt.Sprintf("https://store.steampowered.com/app/%d", dlc.ID),
			})
		}
		if shouldCompactLinkTable(formatter, header, rows) {
			header[2] = app.localize("rules.app_id", "APP ID")
			for index, dlc := range rules.CreatorDLC {
				rows[index][2] = fmt.Sprintf("%d", dlc.ID)
			}
		}

		t := formatter.NewTable(header, rows)
		if err := formatter.PrintTable(t); err != nil {
			return app.wrapError("error.render_rules", "failed to render rules", err)
		}
	}

	// Print Mods
	if len(rules.Mods) > 0 {
		formatter.PrintSectionHeader(app.localize("section.mods", "Mods"))
		header := table.Row{
			app.localize("table.number", "#"),
			app.localize("rules.mod_name", "Mod Name"),
			app.localize("rules.mod_url", "Mod URL"),
		}
		rows := make([]table.Row, 0, len(rules.Mods))

		for i, mod := range rules.Mods {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				mod.Name,
				fmt.Sprintf("https://steamcommunity.com/sharedfiles/filedetails/?id=%d", mod.ID),
			})
		}
		if shouldCompactLinkTable(formatter, header, rows) {
			header[2] = app.localize("rules.workshop_id", "WORKSHOP ID")
			for index, mod := range rules.Mods {
				rows[index][2] = fmt.Sprintf("%d", mod.ID)
			}
		}

		t := formatter.NewTable(header, rows)
		if err := formatter.PrintTable(t); err != nil {
			return err
		}
	}

	// Only print footer message for table format
	if formatter.IsTableFormat() {
		_, _ = fmt.Fprintf(
			app.Out,
			"%s\n",
			app.localize("footer.rules", "A2S_RULES response for %s", address),
		)
	}

	return nil
}

// shouldCompactLinkTable selects the compact human-readable representation
// only when the full link table cannot reach its preferred layout.
func shouldCompactLinkTable(formatter *Formatter, header table.Row, rows []table.Row) bool {
	if formatter == nil || !formatter.IsTableFormat() || formatter.terminalWidth <= 0 {
		return false
	}

	layout, err := planTableLayout(formatter.terminalWidth, header, rows, nil)
	return err == nil && !layout.FitsPreferred
}
