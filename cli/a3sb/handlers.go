package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

// processing A2S_INFO request
func handleInfo(client *a3sb.Client, c *cli.Context) error {
	info, err := client.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	printInfo(info, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))
	return nil
}

// processing A2S_RULES request
func handleRules(client *a3sb.Client, c *cli.Context) error {
	rules, err := client.GetRules(c.Uint64("app-id"))
	if err != nil {
		return fmt.Errorf("failed to get server rules: %w", err)
	}

	printRules(rules, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))
	return nil
}

// processing A2S_PLAYERS request
func handlePlayers(client *a3sb.Client, c *cli.Context) error {
	players, err := client.GetPlayers()
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	printPlayers(players, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))
	return nil
}
