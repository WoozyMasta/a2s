package main

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/tableprinter"
	"github.com/woozymasta/a2s/pkg/a2s"
)

func printPlayers(client *a2s.Client, json bool) {
	players, err := client.GetPlayers()
	if err != nil {
		fatalf("failed to get players: %s", err)
	}

	if json {
		printJSON(players)
	}

	if len(*players) == 0 {
		fmt.Println("The server is empty and there are no players to print ...")
		return
	}

	counter := [4]byte{}
	for _, player := range *players {
		if player.Duration != 0 {
			counter[0]++
		}
		if player.Score != 0 {
			counter[1]++
		}
		if player.Name != "" {
			counter[2]++
		}
		if player.Index != 0 {
			counter[3]++
		}
	}

	columns := []string{"  #"}
	if counter[0] > 0 {
		columns = append(columns, "PlayTime")
	}
	if counter[1] > 0 {
		columns = append(columns, "Score")
	}
	if counter[2] > 0 {
		columns = append(columns, "Name")
	}
	if counter[3] > 0 {
		columns = append(columns, "Index")
	}

	table := tableprinter.NewTablePrinter(columns, "=")

	for i, player := range *players {
		row := []string{fmt.Sprintf("%3d", i+1)}

		if counter[0] > 0 {
			row = append(row, player.Duration.String())
		}
		if counter[1] > 0 {
			row = append(row, fmt.Sprint(player.Score))
		}
		if counter[2] > 0 {
			row = append(row, player.Name)
		}
		if counter[3] > 0 {
			row = append(row, fmt.Sprint(player.Index))
		}

		if err := table.AddRow(row); err != nil {
			fatalf("Create players table: %s", err)
		}
	}

	table.Print()
	fmt.Printf("A2S_PLAYERS response for %s\n", client.Address)
}
