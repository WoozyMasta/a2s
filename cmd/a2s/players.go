package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

// executePlayers queries A2S_PLAYER and renders the available player fields.
func executePlayers(app *Application, cmd *PlayersCommand) error {
	client, err := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	if err != nil {
		return err
	}
	defer closeClient(app, client)

	players, err := client.GetPlayers(context.Background())
	if err != nil {
		return friendlyQueryError("failed to get players", err, client.Timeout())
	}

	formatter := NewFormatter(cmd.Format, app.Out)

	if formatter.ShouldUseJSON() {
		return formatter.PrintJSON(players)
	}

	if len(players) == 0 {
		_, _ = fmt.Fprintln(app.Out, "The server is empty and there are no players to print ...")
		return nil
	}

	// Show only columns containing at least one non-zero value
	// so sparse server responses do not produce empty table columns.
	counter := [4]byte{}
	for _, player := range players {
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
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row(columns))

	for i, player := range players {
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

	if err := formatter.PrintTable(t); err != nil {
		return err
	}

	// Only print footer message for table format
	if formatter.IsTableFormat() {
		_, _ = fmt.Fprintf(app.Out, "A2S_PLAYERS response for %s\n", client.Addr())
	}

	return nil
}
