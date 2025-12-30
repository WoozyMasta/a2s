package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

func executePlayers(cmd *PlayersCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer client.Close()

	players, err := client.GetPlayers()
	if err != nil {
		fatalf("Failed to get players: %s", err)
	}

	formatter := NewFormatter(cmd.Format)

	if formatter.ShouldUseJSON() {
		formatter.PrintJSON(players)
		return
	}

	if len(*players) == 0 {
		fmt.Println("The server is empty and there are no players to print ...")
		return
	}

	// Determine which columns to show
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

	columns := []interface{}{"#"}
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

	t := table.NewWriter()
	if formatter.IsTableFormat() {
		t.SetOutputMirror(os.Stdout)
	}
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row(columns))

	for i, player := range *players {
		row := []interface{}{fmt.Sprintf("%d", i+1)}

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

		t.AppendRow(table.Row(row))
	}

	formatter.PrintTable(t)

	// Only print footer message for table format
	if formatter.IsTableFormat() {
		fmt.Printf("A2S_PLAYERS response for %s\n", client.Address)
	}
}
