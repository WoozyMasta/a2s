package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
	"github.com/woozymasta/a2s/pkg/config"
)

var logFormatter *log.TextFormatter
var host string
var port int

func main() {
	var client *a3sb.Client
	logFormatter = prepareLogging()
	cmd := filepath.Base(os.Args[0])

	app := &cli.App{
		Name:  cmd,
		Usage: "CLI for querying A2S server information and working with A3SB subprotocol for Arma 3 and DayZ",
		UsageText: fmt.Sprintf(
			"%[1]s [global options] command host port\nExample: %[1]s -j info 127.0.0.1 27016", cmd,
		),
		HideHelpCommand: true,
		Flags:           commonFlags(),
		Writer:          os.Stderr,
		Before: func(c *cli.Context) error {
			var err error
			setupLogging(c.String("log-level"), logFormatter)

			if c.Bool("version") {
				fmt.Printf(
					"%s\n\nversion:  %s\ncommit:   %s\nbuilt:    %s\nrepo:     %s\n",
					c.App.Name, config.Version, config.Commit, config.BuildTime, config.URL,
				)
				os.Exit(0)
			}

			if c.Args().Len() < 1 {
				if err := cli.ShowAppHelp(c); err != nil {
					return err
				}
				return fmt.Errorf("command must be provided")
			}

			if c.Args().Len() < 3 {
				addr := strings.Split(c.Args().Get(1), ":")
				if len(addr) == 2 {
					host = addr[0]
					port, err = strconv.Atoi(addr[1])
					if err != nil {
						if err := cli.ShowAppHelp(c); err != nil {
							return err
						}
						return fmt.Errorf("invalid port %s", c.Args().Get(2))
					}
				} else {
					return fmt.Errorf("host and port must be provided as positional arguments")
				}
			} else if c.Args().Len() > 2 {
				host = c.Args().Get(1)
				port, err = strconv.Atoi(c.Args().Get(2))
				if err != nil {
					if err := cli.ShowAppHelp(c); err != nil {
						return err
					}
					return fmt.Errorf("invalid port %s", c.Args().Get(2))
				}
			}

			clientA2S, err := createClient(host, port, c)
			if err != nil {
				return err
			}
			client = &a3sb.Client{Client: clientA2S}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "info",
				Usage: "Retrieve server information A3SB_INFO",
				Action: func(c *cli.Context) error {
					return handleInfo(client, c)
				},
			},
			{
				Name:  "rules",
				Usage: "Retrieve server rules A3SB_RULES",
				Action: func(c *cli.Context) error {
					return handleRules(client, c)
				},
			},
			{
				Name:  "players",
				Usage: "Retrieve player list A3SB_PLAYERS",
				Action: func(c *cli.Context) error {
					return handlePlayers(client, c)
				},
			},
			{
				Name:  "all",
				Usage: "Retrieve all available server information",
				Action: func(c *cli.Context) error {
					if err := handleInfo(client, c); err != nil {
						return err
					}
					if err := handleRules(client, c); err != nil {
						return err
					}
					return handlePlayers(client, c)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func commonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			EnvVars: []string{"A3SB_JSON"},
			Usage:   "output in JSON format",
		},
		&cli.Uint64Flag{
			Name:    "app-id",
			Aliases: []string{"i"},
			EnvVars: []string{"A3SB_APP_ID"},
			Usage:   "appID DayZ (221100) or Arma 3 (107410) for more accurate results",
			Value:   0,
		},
		&cli.IntFlag{
			Name:    "deadline-timeout",
			Aliases: []string{"t"},
			EnvVars: []string{"A3SB_DEADLINE_TIMEOUT"},
			Usage:   "set timeout in seconds",
			Value:   int(a2s.DefaultDeadlineTimeout),
		},
		&cli.IntFlag{
			Name:    "buffer-size",
			Aliases: []string{"b"},
			EnvVars: []string{"A3SB_BUFFER_SIZE"},
			Usage:   "set buffer size",
			Value:   8192,
		},
		&cli.StringFlag{
			Name:    "log-level",
			Value:   "error",
			Usage:   "set log level",
			Aliases: []string{"l"},
			EnvVars: []string{"A3SB_LOG_LEVEL"},
		},
		&cli.BoolFlag{
			Name:               "version",
			Aliases:            []string{"v"},
			Usage:              "print version",
			DisableDefaultText: true,
		},
	}
}
