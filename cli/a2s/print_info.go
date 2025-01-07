package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/tableprinter"
)

func printInfo(info *a2s.Info, address string, json bool) {
	if json {
		printJson(info)
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
		log.Fatalf("Create info table: %s", err)
	}

	// GoldSource
	if info.Format == 0x6D {
		if err := table.AddRow([]string{"Server address:", info.Address}); err != nil {
			log.Fatalf("Create info table (GoldSource address): %s", err)
		}

		if info.Mod != nil {
			if err := table.AddRows([][]string{
				{"Mod URL:", info.Mod.Link},
				{"Download URL:", info.Mod.DownloadLink},
				{"Mod Version:", fmt.Sprintf("%d", info.Mod.Version)},
				{"Mod Size:", fmt.Sprintf("%d", info.Mod.Size)},
				{"Multiplayer only:", fmt.Sprintf("%t", info.Mod.Type)},
				{"Custom DLL:", fmt.Sprintf("%t", info.Mod.DLL)},
			}); err != nil {
				log.Fatalf("Create info table (GoldSource Mod): %s", err)
			}
		}
	}

	if info.EDF != 0 {
		if info.Port != 0 {
			if err := table.AddRow([]string{"Port:", fmt.Sprintf("%d", info.Port)}); err != nil {
				log.Fatalf("Create info table (EDF Port): %s", err)
			}
		}

		if info.SteamID != 0 {
			if err := table.AddRow([]string{"Server SteamID:", fmt.Sprintf("%d", info.SteamID)}); err != nil {
				log.Fatalf("Create info table (EDF ServerID): %s", err)
			}
		}

		if (info.EDF & 0x40) != 0 {
			if err := table.AddRows([][]string{
				{"SourceTV Port:", fmt.Sprintf("%d", info.SourceTVPort)},
				{"SourceTV Name:", info.SourceTVName},
			}); err != nil {
				log.Fatalf("Create info table (EDF SourceTV): %s", err)
			}
		}

		if len(info.Keywords) > 0 {
			max := len(info.Name)
			if 60 > max {
				max = 60
			}

			for i, k := range tableprinter.JoinWithLimit(info.Keywords, ", ", max) {
				if i == 0 {
					if err := table.AddRow([]string{"Keywords:", k}); err != nil {
						log.Fatalf("Create info table (keywords key): %s", err)
					}
				} else {
					if err := table.AddRow([]string{"", k}); err != nil {
						log.Fatalf("Create info table (keywords value): %s", err)
					}
				}
			}
		}
	}

	if err := table.AddRow([]string{"Server ping:", fmt.Sprintf("%d ms", info.Ping.Milliseconds())}); err != nil {
		log.Fatalf("Create info table (ping): %s", err)
	}

	return table
}
