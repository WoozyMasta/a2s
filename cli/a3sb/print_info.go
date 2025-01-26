package main

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/tableprinter"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords"
	"github.com/woozymasta/steam/utils/appid"
)

func printInfo(client *a2s.Client, json bool) {
	info, err := client.GetInfo()
	if err != nil {
		fatalf("Failed to get server info: %s", err)
	}

	if json {
		printKeywordsJSON(info)
		return
	}

	table := tableprinter.NewTablePrinter([]string{"Property", "Value"}, "=")

	if err := table.AddRows([][]string{
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
	}); err != nil {
		fatalf("Create info table: %s", err)
	}

	if info.Port != 0 {
		if err := table.AddRow([]string{"Port:", fmt.Sprintf("%d", info.Port)}); err != nil {
			fatalf("Create info table (EDF Port): %s", err)
		}
	}

	if info.SteamID != 0 {
		if err := table.AddRow([]string{"Server SteamID:", fmt.Sprintf("%d", info.SteamID)}); err != nil {
			fatalf("Create info table (EDF ServerID): %s", err)
		}
	}

	switch info.ID {
	case appid.Arma3.Uint64():
		arma := keywords.ParseArma3(info.Keywords)
		if err := table.AddRows([][]string{
			{"Type of game:", arma.GameType.String()},
			{"Server OS:", arma.Platform.String()},
			{"Content hash:", arma.LoadedContentHash},
			{"Country:", arma.Country},
			{"Island name:", arma.Island},
			{"Time left:", fmt.Sprintf("%s", arma.TimeLeft)},
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
			{"Enabled fle patching:", fmt.Sprintf("%t", arma.AllowedFilePatching)},
		}); err != nil {
			fatalf("Create info table (DayZ keywords): %s", err)
		}

	case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
		dayz := keywords.ParseDayZ(info.Keywords)
		if err := table.AddRows([][]string{
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
			{"Fle patching:", fmt.Sprintf("%t", dayz.FlePatching)},
			{"Need DLC:", fmt.Sprintf("%t", dayz.DLC)},
		}); err != nil {
			fatalf("Create info table (DayZ keywords): %s", err)
		}
	}

	if err := table.AddRow([]string{"Server ping:", fmt.Sprintf("%d ms", info.Ping.Milliseconds())}); err != nil {
		fatalf("Create info table (ping): %s", err)
	}

	table.Print()
	fmt.Printf("A2S_INFO response for %s\n", client.Address)
}
