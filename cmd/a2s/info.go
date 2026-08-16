// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
	"github.com/woozymasta/a2s/pkg/keywords"
)

// executeInfo queries A2S_INFO and renders the result in the requested format.
func executeInfo(app *Application, cmd *InfoCommand) error {
	client, err := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	if err != nil {
		return app.wrapError("error.client_create", "failed to create client", err)
	}
	defer closeClient(app, client)

	info, meta, err := client.GetInfoWithMeta(context.Background())
	if err != nil {
		return friendlyQueryError(app, "error.server_info", "failed to get server info", err, client.Timeout())
	}

	formatter := NewFormatter(cmd.Format, app.Out, app.Localizer)

	if formatter.ShouldUseJSON() {
		return printInfoJSON(info, formatter)
	}

	return renderInfoTable(app, info, meta, client.Addr().String(), cmd.Format, formatter)
}

// renderInfoTable renders human-readable INFO fields without querying a server.
func renderInfoTable(
	app *Application,
	info *a2s.Info,
	meta a2s.QueryMeta,
	address, format string,
	formatter *Formatter,
) error {
	// Human-readable formats share the table assembly below.
	header := table.Row{
		app.localize("table.property", "Property"),
		app.localize("table.value", "Value"),
	}
	rows := make([]table.Row, 0, 48)

	rows = append(rows, []table.Row{
		{
			app.localize("info.query_type", "Query type:"),
			info.Format.String(),
		},
		{
			app.localize("info.protocol", "Protocol:"),
			fmt.Sprintf("%d", info.Protocol),
		},
		{
			app.localize("info.server_name", "Server name:"),
			info.Name,
		},
		{
			app.localize("info.map", "Map on server:"),
			info.Map,
		},
		{
			app.localize("info.game_folder", "Game folder:"),
			info.Folder,
		},
		{
			app.localize("info.game_name", "Game name:"),
			info.Game,
		},
		{
			app.localize("info.game_id", "Game ID:"),
			formatAppID(info.EffectiveID()),
		},
		{
			app.localize("info.players_slots", "Players/Slots:"),
			fmt.Sprintf("%d/%d", info.Players, info.MaxPlayers),
		},
		{
			app.localize("info.bots_count", "Bots count:"),
			fmt.Sprintf("%d", info.Bots),
		},
		{
			app.localize("info.server_type", "Server type:"),
			info.ServerType.String(),
		},
		{
			app.localize("info.server_os", "Server OS:"),
			info.Environment.String(),
		},
		{
			app.localize("info.need_password", "Need password:"),
			app.formatBool(info.Visibility),
		},
		{
			app.localize("info.vac_protected", "VAC protected:"),
			app.formatBool(info.VAC),
		},
		{
			app.localize("info.game_version", "Game version:"),
			info.Version,
		},
	}...)

	// GoldSource fields are only meaningful for the obsolete GoldSource layout.
	if info.Format == 0x6D {
		if info.Address != "" {
			rows = append(rows, table.Row{
				app.localize("info.server_address", "Server address:"),
				info.Address,
			})
		}

		if info.Mod != nil {
			rows = append(rows, []table.Row{
				{
					app.localize("info.mod_url", "Mod URL:"),
					info.Mod.Link,
				},
				{
					app.localize("info.download_url", "Download URL:"),
					info.Mod.DownloadLink,
				},
				{
					app.localize("info.mod_version", "Mod Version:"),
					fmt.Sprintf("%d", info.Mod.Version),
				},
				{
					app.localize("info.mod_size", "Mod Size:"),
					fmt.Sprintf("%d", info.Mod.Size),
				},
				{
					app.localize("info.multiplayer_only", "Multiplayer only:"),
					fmt.Sprintf("%t", info.Mod.Type),
				},
				{
					app.localize("info.custom_dll", "Custom DLL:"),
					fmt.Sprintf("%t", info.Mod.DLL),
				},
			}...)
		}
	}

	// Render only optional fields advertised by EDF.
	if info.EDF != 0 {
		if info.Port != 0 {
			rows = append(rows, table.Row{
				app.localize("info.port", "Port:"),
				fmt.Sprintf("%d", info.Port),
			})
		}

		if info.SteamID != 0 {
			rows = append(rows, table.Row{
				app.localize("info.server_steamid", "Server SteamID:"),
				fmt.Sprintf("%d", info.SteamID),
			})
		}

		if (info.EDF & 0x40) != 0 {
			rows = append(rows, []table.Row{
				{
					app.localize("info.sourcetv_port", "SourceTV Port:"),
					fmt.Sprintf("%d", info.SourceTVPort),
				},
				{
					app.localize("info.sourcetv_name", "SourceTV Name:"),
					info.SourceTVName,
				},
			}...)
		}

		if len(info.Keywords) > 0 {
			// Keywords have game-specific structure only
			// for the supported Arma 3 and DayZ AppIDs;
			// other games keep the raw list above.
			switch info.EffectiveID() {
			case appid.Arma3:
				arma := keywords.ParseArma3(info.Keywords)
				rows = append(rows, []table.Row{
					{
						app.localize("info.type_of_game", "Type of game:"),
						arma.GameType.String(),
					},
					{
						app.localize("info.server_os", "Server OS:"),
						arma.Platform.String(),
					},
					{
						app.localize("info.content_hash", "Content hash:"),
						arma.LoadedContentHash,
					},
					{
						app.localize("info.country", "Country:"),
						arma.Country,
					},
					{
						app.localize("info.island_name", "Island name:"),
						arma.Island,
					},
					{
						app.localize("info.time_left", "Time left:"),
						arma.TimeLeft.String(),
					},
					{
						app.localize("info.required_version", "Required version:"),
						fmt.Sprintf("%d", arma.RequiredVersion),
					},
					{
						app.localize("info.required_build", "Required build:"),
						fmt.Sprintf("%d", arma.RequiredBuildNo),
					},
					{
						app.localize("info.language", "Language:"),
						arma.Language.String(),
					},
					{
						app.localize("info.longitude", "Longitude:"),
						fmt.Sprintf("%d", arma.Longitude),
					},
					{
						app.localize("info.latitude", "Latitude:"),
						fmt.Sprintf("%d", arma.Latitude),
					},
					{
						app.localize("info.state_of_server", "State of server:"),
						arma.ServerState.String(),
					},
					{
						app.localize("info.battleye_protected", "BattlEye protected:"),
						app.formatBool(arma.BattlEye),
					},
					{
						app.localize("info.difficulty", "Difficulty:"),
						fmt.Sprintf("%d", arma.Difficulty),
					},
					{
						app.localize("info.require_mods_equal", "Require mods equal:"),
						app.formatBool(arma.EqualModRequired),
					},
					{
						app.localize("info.locked_state", "Locked state:"),
						app.formatBool(arma.Lock),
					},
					{
						app.localize("info.verify_signatures", "Verify signatures:"),
						app.formatBool(arma.VerifySignatures),
					},
					{
						app.localize("info.dedicated", "Dedicated:"),
						app.formatBool(arma.Dedicated),
					},
					{
						app.localize("info.enabled_file_patching", "Enabled file patching:"),
						app.formatBool(arma.AllowedFilePatching),
					},
				}...)

			case appid.DayZ, appid.DayZExperimental:
				dayz := keywords.ParseDayZ(info.Keywords)
				rows = append(rows, []table.Row{
					{
						app.localize("info.shard", "Shard:"),
						dayz.Shard,
					},
					{
						app.localize("info.in_game_time", "In game time:"),
						dayz.Time.String(),
					},
					{
						app.localize("info.time_day_x", "Time day x:"),
						fmt.Sprintf("%f", dayz.TimeDayAccel),
					},
					{
						app.localize("info.time_night_x", "Time night x:"),
						fmt.Sprintf("%f", dayz.TimeNightAccel),
					},
					{
						app.localize("info.game_port", "Game port:"),
						fmt.Sprintf("%d", dayz.GamePort),
					},
					{
						app.localize("info.players_queue", "Players queue:"),
						fmt.Sprintf("%d", dayz.PlayersQueue),
					},
					{
						app.localize("info.battleye_protected", "BattlEye protected:"),
						app.formatBool(dayz.BattlEye),
					},
					{
						app.localize("info.third_person", "Third person:"),
						app.formatBool(!dayz.NoThirdPerson),
					},
					{
						app.localize("info.external", "External:"),
						app.formatBool(dayz.External),
					},
					{
						app.localize("info.private_hive", "Private hive:"),
						app.formatBool(dayz.PrivateHive),
					},
					{
						app.localize("info.modded", "Modded:"),
						app.formatBool(dayz.Modded),
					},
					{
						app.localize("info.whitelist", "Whitelist:"),
						app.formatBool(dayz.Whitelist),
					},
					{
						app.localize("info.file_patching", "File patching:"),
						app.formatBool(dayz.FlePatching),
					},
					{
						app.localize("info.need_dlc", "Need DLC:"),
						app.formatBool(dayz.DLC),
					},
				}...)
			}
		}
	}

	rows = append(rows, table.Row{
		app.localize("info.server_ping", "Server ping:"),
		fmt.Sprintf("%d ms", meta.Duration.Milliseconds()),
	})

	t := formatter.NewTable(header, rows)
	if err := formatter.PrintTable(t); err != nil {
		return app.wrapError("error.render_info", "failed to render server info", err)
	}

	// Only print footer message for table format
	if format == "table" || format == "" {
		_, _ = fmt.Fprintf(
			app.Out,
			"%s\n",
			app.localize("footer.info", "A2S_INFO response for %s", address),
		)
	}

	return nil
}

// formatAppID returns a known game name with its numeric effective ID,
// or only the original numeric value when the ID is unknown.
func formatAppID(id uint64) string {
	steamID := appid.AppID(id)
	if name, ok := steamID.Name(); ok {
		return fmt.Sprintf("%s (%d)", name, id)
	}

	return steamID.String()
}

// printInfoJSON preserves raw keywords for generic games and replaces them
// with the existing typed representation for the supported Arma 3 and DayZ formats.
// The replacement is intentional for those two game-specific schemas;
// unsupported games keep the wire-level keyword list unchanged.
func printInfoJSON(info *a2s.Info, formatter *Formatter) error {
	// Create a map to hold the JSON structure
	jsonMap := make(map[string]any)

	// Marshal info to JSON first
	jsonData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			localizeFormatterError(
				formatter.localizer,
				"error.info_json_marshal",
				"failed to marshal Info",
			),
			err,
		)
	}

	// Unmarshal into a map to add custom fields
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return fmt.Errorf(
			"%s: %w",
			localizeFormatterError(
				formatter.localizer,
				"error.info_json_unmarshal",
				"failed to unmarshal Info JSON",
			),
			err,
		)
	}

	switch info.EffectiveID() {
	case appid.Arma3:
		delete(jsonMap, "keywords")
		armaData := keywords.ParseArma3(info.Keywords)
		jsonMap["keywords"] = armaData

	case appid.DayZ, appid.DayZExperimental:
		delete(jsonMap, "keywords")
		dayZData := keywords.ParseDayZ(info.Keywords)
		jsonMap["keywords"] = dayZData
	}

	return formatter.PrintJSON(jsonMap)
}
