package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
	"github.com/woozymasta/a2s/pkg/keywords"
)

// executeInfo queries A2S_INFO and renders the result in the requested format.
func executeInfo(cmd *InfoCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer closeClient(client)

	info, meta, err := client.GetInfoWithMeta(context.Background())
	if err != nil {
		fatalf("Failed to get server info: %s", err)
	}

	formatter := NewFormatter(cmd.Format)

	if formatter.ShouldUseJSON() {
		printInfoJSON(info, formatter)
		return
	}

	// JSON output uses Info's own schema and does not need table-specific fields.
	// All other formats share the table assembly below.
	t := table.NewWriter()
	if formatter.IsTableFormat() {
		t.SetOutputMirror(os.Stdout)
	}
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Property", "Value"})

	t.AppendRows([]table.Row{
		{"Query type:", info.Format.String()},
		{"Protocol:", fmt.Sprintf("%d", info.Protocol)},
		{"Server name:", info.Name},
		{"Map on server:", info.Map},
		{"Game folder:", info.Folder},
		{"Game name:", info.Game},
		{"Steam AppID:", formatAppID(info.ID)},
		{"Players/Slots:", fmt.Sprintf("%d/%d", info.Players, info.MaxPlayers)},
		{"Bots count:", fmt.Sprintf("%d", info.Bots)},
		{"Server type:", info.ServerType.String()},
		{"Server OS:", info.Environment.String()},
		{"Need password:", fmt.Sprintf("%t", info.Visibility)},
		{"VAC protected:", fmt.Sprintf("%t", info.VAC)},
		{"Game version:", info.Version},
	})

	// GoldSource fields are only meaningful for the obsolete GoldSource layout.
	if info.Format == 0x6D {
		if info.Address != "" {
			t.AppendRow(table.Row{"Server address:", info.Address})
		}

		if info.Mod != nil {
			t.AppendRows([]table.Row{
				{"Mod URL:", info.Mod.Link},
				{"Download URL:", info.Mod.DownloadLink},
				{"Mod Version:", fmt.Sprintf("%d", info.Mod.Version)},
				{"Mod Size:", fmt.Sprintf("%d", info.Mod.Size)},
				{"Multiplayer only:", fmt.Sprintf("%t", info.Mod.Type)},
				{"Custom DLL:", fmt.Sprintf("%t", info.Mod.DLL)},
			})
		}
	}

	// Render only optional fields advertised by EDF.
	if info.EDF != 0 {
		if info.Port != 0 {
			t.AppendRow(table.Row{"Port:", fmt.Sprintf("%d", info.Port)})
		}

		if info.SteamID != 0 {
			t.AppendRow(table.Row{"Server SteamID:", fmt.Sprintf("%d", info.SteamID)})
		}

		if (info.EDF & 0x40) != 0 {
			t.AppendRows([]table.Row{
				{"SourceTV Port:", fmt.Sprintf("%d", info.SourceTVPort)},
				{"SourceTV Name:", info.SourceTVName},
			})
		}

		if len(info.Keywords) > 0 {
			// Keywords have game-specific structure only
			// for the supported Arma 3 and DayZ AppIDs;
			// other games keep the raw list above.
			switch info.ID {
			case appid.Arma3:
				arma := keywords.ParseArma3(info.Keywords)
				t.AppendRows([]table.Row{
					{"Type of game:", arma.GameType.String()},
					{"Server OS:", arma.Platform.String()},
					{"Content hash:", arma.LoadedContentHash},
					{"Country:", arma.Country},
					{"Island name:", arma.Island},
					{"Time left:", arma.TimeLeft.String()},
					{"Required version:", fmt.Sprintf("%d", arma.RequiredVersion)},
					{"Required build:", fmt.Sprintf("%d", arma.RequiredBuildNo)},
					{"Language:", arma.Language.String()},
					{"Longitude:", fmt.Sprintf("%d", arma.Longitude)},
					{"Latitude:", fmt.Sprintf("%d", arma.Latitude)},
					{"State of server:", arma.ServerState.String()},
					{"BattlEye protected:", fmt.Sprintf("%t", arma.BattlEye)},
					{"Difficulty:", fmt.Sprintf("%d", arma.Difficulty)},
					{"Require mods equal:", fmt.Sprintf("%t", arma.EqualModRequired)},
					{"Locked state:", fmt.Sprintf("%t", arma.Lock)},
					{"Verify signatures:", fmt.Sprintf("%t", arma.VerifySignatures)},
					{"Dedicated:", fmt.Sprintf("%t", arma.Dedicated)},
					{"Enabled file patching:", fmt.Sprintf("%t", arma.AllowedFilePatching)},
				})

			case appid.DayZ, appid.DayZExperimental:
				dayz := keywords.ParseDayZ(info.Keywords)
				t.AppendRows([]table.Row{
					{"Shard:", dayz.Shard},
					{"In game time:", dayz.Time.String()},
					{"Time day x:", fmt.Sprintf("%f", dayz.TimeDayAccel)},
					{"Time night x:", fmt.Sprintf("%f", dayz.TimeNightAccel)},
					{"Game port:", fmt.Sprintf("%d", dayz.GamePort)},
					{"Players queue:", fmt.Sprintf("%d", dayz.PlayersQueue)},
					{"BattlEye protected:", fmt.Sprintf("%t", dayz.BattlEye)},
					{"Third person:", fmt.Sprintf("%t", !dayz.NoThirdPerson)},
					{"External:", fmt.Sprintf("%t", dayz.External)},
					{"Private hive:", fmt.Sprintf("%t", dayz.PrivateHive)},
					{"Modded:", fmt.Sprintf("%t", dayz.Modded)},
					{"Whitelist:", fmt.Sprintf("%t", dayz.Whitelist)},
					{"File patching:", fmt.Sprintf("%t", dayz.FlePatching)},
					{"Need DLC:", fmt.Sprintf("%t", dayz.DLC)},
				})
			}
		}
	}

	t.AppendRow(table.Row{"Server ping:", fmt.Sprintf("%d ms", meta.Duration.Milliseconds())})

	formatter.PrintTable(t)

	// Only print footer message for table format
	if cmd.Format == "table" || cmd.Format == "" {
		fmt.Printf("A2S_INFO response for %s\n", client.Addr())
	}
}

// formatAppID returns a known game name with its numeric AppID,
// or only the original numeric value when the ID is unknown or does not fit AppID.
func formatAppID(id uint64) string {
	if id > uint64(^uint32(0)) {
		return fmt.Sprintf("%d", id)
	}

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
func printInfoJSON(info *a2s.Info, formatter *Formatter) {
	// Create a map to hold the JSON structure
	jsonMap := make(map[string]any)

	// Marshal info to JSON first
	jsonData, err := json.Marshal(info)
	if err != nil {
		fatalf("Failed to marshal Info: %v", err)
	}

	// Unmarshal into a map to add custom fields
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		fatalf("Failed to unmarshal JSON: %v", err)
	}

	switch info.ID {
	case appid.Arma3:
		delete(jsonMap, "keywords")
		armaData := keywords.ParseArma3(info.Keywords)
		jsonMap["keywords"] = armaData

	case appid.DayZ, appid.DayZExperimental:
		delete(jsonMap, "keywords")
		dayZData := keywords.ParseDayZ(info.Keywords)
		jsonMap["keywords"] = dayZData
	}

	formatter.PrintJSON(jsonMap)
}
