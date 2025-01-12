package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords"
	"github.com/woozymasta/a2s/pkg/tableprinter"
)

func printInfo(info *a2s.Info, address string, json bool) {
	if json {
		printJSONWithDayZ(info)
	} else {
		table := makeInfo(info)
		table.Print()
		fmt.Printf("A2S_INFO response for %s\n", address)
	}
}

func makeInfo(info *a2s.Info) *tableprinter.TablePrinter {
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
		log.Fatal().Msgf("Create info table: %s", err)
	}

	if info.Port != 0 {
		if err := table.AddRow([]string{"Port:", fmt.Sprintf("%d", info.Port)}); err != nil {
			log.Fatal().Msgf("Create info table (EDF Port): %s", err)
		}
	}

	if info.SteamID != 0 {
		if err := table.AddRow([]string{"Server SteamID:", fmt.Sprintf("%d", info.SteamID)}); err != nil {
			log.Fatal().Msgf("Create info table (EDF ServerID): %s", err)
		}
	}

	dayz := keywords.ParseDayZ(info.Keywords)
	if dayz.Shard != "" {
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
			log.Fatal().Msgf("Create info table (DayZ keywords): %s", err)
		}
	}

	if err := table.AddRow([]string{"Server ping:", fmt.Sprintf("%d ms", info.Ping.Milliseconds())}); err != nil {
		log.Fatal().Msgf("Create info table (ping): %s", err)
	}

	return table
}
