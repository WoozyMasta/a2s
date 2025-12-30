package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords"
	"github.com/woozymasta/steam/utils/appid"
)

func executeInfo(cmd *InfoCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer closeClient(client)

	info, err := client.GetInfo()
	if err != nil {
		fatalf("Failed to get server info: %s", err)
	}

	formatter := NewFormatter(cmd.Format)

	if formatter.ShouldUseJSON() {
		printInfoJSON(info, formatter)
		return
	}

	// Table output
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
		{"Steam AppID:", fmt.Sprintf("%d", info.ID)},
		{"Players/Slots:", fmt.Sprintf("%d/%d", info.Players, info.MaxPlayers)},
		{"Bots count:", fmt.Sprintf("%d", info.Bots)},
		{"Server type:", info.ServerType.String()},
		{"Server OS:", info.Environment.String()},
		{"Need password:", fmt.Sprintf("%t", info.Visibility)},
		{"VAC protected:", fmt.Sprintf("%t", info.VAC)},
		{"Game version:", info.Version},
	})

	// GoldSource specific fields
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

	// EDF fields
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
			// Parse keywords for Arma3/DayZ
			switch info.ID {
			case appid.Arma3.Uint64():
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

			case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
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

	t.AppendRow(table.Row{"Server ping:", fmt.Sprintf("%d ms", info.Ping.Milliseconds())})

	formatter.PrintTable(t)

	// Only print footer message for table format
	if cmd.Format == "table" || cmd.Format == "" {
		fmt.Printf("A2S_INFO response for %s\n", client.Address)
	}
}

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

	// Parse keywords for Arma3/DayZ
	delete(jsonMap, "keywords")

	switch info.ID {
	case appid.Arma3.Uint64():
		armaData := keywords.ParseArma3(info.Keywords)
		jsonMap["keywords"] = armaData
	case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
		dayZData := keywords.ParseDayZ(info.Keywords)
		jsonMap["keywords"] = dayZData
	}

	formatter.PrintJSON(jsonMap)
}
