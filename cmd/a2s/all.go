package main

import (
	"fmt"
)

func executeAll(cmd *AllCommand) {
	if cmd.Args.Host == "" {
		fatal("Host must be provided")
	}

	client := createClient(cmd.Args.Host, cmd.Args.Port, cmd.Timeout, cmd.Buffer)
	defer client.Close()

	// Execute info
	infoCmd := InfoCommand{
		GlobalOptions: cmd.GlobalOptions,
	}
	infoCmd.Args.Host = cmd.Args.Host
	infoCmd.Args.Port = cmd.Args.Port
	executeInfo(&infoCmd)

	fmt.Println()

	// Execute rules
	rulesCmd := RulesCommand{
		GlobalOptions: cmd.GlobalOptions,
		RulesOptions:  cmd.RulesOptions,
	}
	rulesCmd.Args.Host = cmd.Args.Host
	rulesCmd.Args.Port = cmd.Args.Port
	executeRules(&rulesCmd)

	fmt.Println()

	// Execute players
	playersCmd := PlayersCommand{
		GlobalOptions: cmd.GlobalOptions,
	}
	playersCmd.Args.Host = cmd.Args.Host
	playersCmd.Args.Port = cmd.Args.Port
	executePlayers(&playersCmd)
}
