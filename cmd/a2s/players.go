package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/woozymasta/a2s/pkg/a2s"
)

// executePlayers queries A2S_PLAYER and renders the available player fields.
func executePlayers(app *Application, cmd *PlayersCommand) error {
	client, err := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	if err != nil {
		return app.wrapError("error.client_create", "failed to create client", err)
	}
	defer closeClient(app, client)

	players, err := client.GetPlayers(context.Background())
	if err != nil {
		return friendlyQueryError(app, "error.players", "failed to get players", err, client.Timeout())
	}

	formatter := NewFormatter(cmd.Format, app.Out, app.Localizer)

	if formatter.ShouldUseJSON() {
		return formatter.PrintJSON(players)
	}

	return renderPlayersTable(app, players, client.Addr().String(), formatter)
}

// renderPlayersTable renders human-readable player data without querying a server.
func renderPlayersTable(app *Application, players []a2s.Player, address string, formatter *Formatter) error {
	if len(players) == 0 {
		_, _ = fmt.Fprintln(app.Out, app.localize(
			"players.empty",
			"The server is empty and there are no players to print ...",
		))
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

	columns := []interface{}{app.localize("table.number", "#")}
	if counter[0] > 0 {
		columns = append(columns, app.localize("players.play_time", "PlayTime"))
	}
	if counter[1] > 0 {
		columns = append(columns, app.localize("players.score", "Score"))
	}
	if counter[2] > 0 {
		columns = append(columns, app.localize("players.name", "Name"))
	}
	if counter[3] > 0 {
		columns = append(columns, app.localize("players.index", "Index"))
	}

	rows := make([]table.Row, 0, len(players))
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

		rows = append(rows, table.Row(row))
	}

	t := formatter.NewTable(table.Row(columns), rows)
	if err := formatter.PrintTable(t); err != nil {
		return app.wrapError("error.render_players", "failed to render players", err)
	}

	// Only print footer message for table format
	if formatter.IsTableFormat() {
		_, _ = fmt.Fprintf(
			app.Out,
			"%s\n",
			app.localize("footer.players", "A2S_PLAYERS response for %s", address),
		)
	}

	return nil
}
